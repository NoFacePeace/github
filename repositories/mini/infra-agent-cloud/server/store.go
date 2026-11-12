package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store 定义任务创建流程所需的数据存取接口。
type Store interface {
	GetAgent(ctx context.Context, agentID string) (*Agent, error)
	GetVersion(ctx context.Context, versionID string) (*AgentVersion, error)
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	// GetLastSession 返回任务的上一轮 Session，无记录时返回 nil, nil。
	// 实现须按稳定的执行轮次顺序取最新一条，不能按随机 ID 判断先后。
	GetLastSession(ctx context.Context, taskID string) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	// StopSession 在数据库事务中锁定最新 Session，将 running 改为 stopped 并返回最新快照。
	// 重复停止须幂等，不覆盖 completed、failed 等已有终态；不存在时返回 nil, nil。
	// 不释放执行租约或容量；非 running 状态须拒绝后续执行续约。
	StopSession(ctx context.Context, sessionID string) (*Session, error)
	// UpdateSessionDispatch 在数据库事务内锁定并读取最新 Session，返回更新后的快照。
	// 仅对 running 会话更新 Dispatch；已有更晚阶段的摘要时，忽略迟到的 queued 摘要。
	// failed 调度将 running 会话置为 failed 并记录错误，cancelled 调度置为 stopped。
	// 不覆盖已有业务终态，不更新其他会话字段；事务提交后释放行锁，再由 Service 做远程补偿。
	UpdateSessionDispatch(ctx context.Context, sessionID string, dispatch *SessionDispatch) (*Session, error)
	// UpdateSessionRunner 仅更新 Session 的 RunnerID，不修改状态及其他字段。
	// 相同绑定重复更新须幂等；已有不同 Runner 绑定时返回冲突，不得直接覆盖。
	UpdateSessionRunner(ctx context.Context, sessionID, runnerID string) error
	// ListRecoverableSessions 分页查询恢复候选：早于 startupBefore 创建且尚未启动执行的 running 会话、
	// 执行租约已失效的 running 会话，以及已 failed 或 stopped 但仍持有有效执行租约的会话。
	// 租约有效性以存储端当前时间判断；候选仅供 Service 复查，不代表可直接更新状态。
	// cursor 为空表示第一页，limit 须为正数；使用稳定游标分页，返回的下一页游标为空表示扫描结束。
	// Service 须结合 Job 调度状态和截止时间复查，不能仅因尚无执行租约就判定排队会话失败。
	ListRecoverableSessions(ctx context.Context, startupBefore time.Time, cursor string, limit int) ([]Session, string, error)
	// FailSessionIfLeaseExpired 在数据库事务内锁定并复查最新 Session、Job 阶段和 Redis 执行租约。
	// 仅当 Session 仍 running、CreatedAt 非零且不晚于 startupBefore，Job 不存在或已 dispatched，
	// 并且执行租约仍不存在时，才标记 failed 并保存原因；不覆盖已有终态，不释放租约或容量。
	// 更新成功返回 true；复查条件不满足返回 false, nil；存储异常返回错误。
	FailSessionIfLeaseExpired(ctx context.Context, sessionID string, startupBefore time.Time, reason string) (bool, error)
	// CreateSession 创建即为 running，须通过数据库唯一约束保证同一 Task 最多有一个 running 的 Session。
	// MySQL 可使用生成列：running 时取 TaskID，其他状态取 NULL，再对该列建立唯一索引。
	// 历史会话不受影响；并发创建触发唯一键冲突时，应返回任务已有活跃会话的错误。
	// 状态更新进入活跃状态时同样受此约束保护，不能仅依赖 Service 先查询后创建。
	CreateSession(ctx context.Context, session *Session) error
	// FailSession 将未结束的 Session 原子标记为 failed 并保存原因，不覆盖已有终态。
	// 失败状态作为撤销执行授权的依据，后续预留和执行续约须检查并拒绝该 Session。
	// 不立即清除执行 token、执行租约期限或释放容量，等待退出确认或原租约到期后回收。
	// 重复调用不改变已记录的终态；本方法不发送停止指令，也不保证远程进程已退出。
	FailSession(ctx context.Context, sessionID, reason string) error
	// EnqueueJob 按 SessionID 原子去重，首次提交时保存 Job 并加入待调度队列。
	// SessionID 对应的 Job 已存在时直接成功，不重复入队，也不覆盖已有 Job。
	// Redis 实现通过 Lua 原子保存记录和队列成员；数据库实现须使用 SessionID 唯一约束。
	EnqueueJob(ctx context.Context, job *Job) error
	// ListPendingJobIDs 按入队顺序返回待调度 Job 的标识，当前使用 SessionID。
	ListPendingJobIDs(ctx context.Context) ([]string, error)
	// ClaimJob 获取指定 Job 的独立调度锁，未获取到时返回空 token 和 nil。
	// ttl 仅控制调度锁的有效期，不设置 Job 的过期时间，也不移除队列成员。
	ClaimJob(ctx context.Context, sessionID string, ttl time.Duration) (string, error)
	// GetJob 按 SessionID 查询最新 Job，包含调度状态和执行归属；不存在时返回 nil, nil。
	// 调度流程先获取锁再读取，其他查询场景无需获取调度锁。
	GetJob(ctx context.Context, sessionID string) (*Job, error)
	// SaveJob 校验独立调度锁的 token 后保存完整 Job，状态由调用方设置。
	// 准备完成时将状态设为 dispatching，并保存固定消息，供后续投递重试复用。
	// 保存成功返回 true；锁不存在或 token 不匹配返回 false, nil，不覆盖记录。
	// 不修改执行租约、Session 或待调度集合，也不释放调度锁。
	SaveJob(ctx context.Context, job *Job, claimToken string) (bool, error)
	// CancelQueuedJob 原子将 queued 改为 cancelled，并删除调度锁，阻止旧调度器继续保存。
	// 已 cancelled 返回 true；记录不存在或已进入其他状态返回 false，不覆盖其状态。
	// 保留调度记录和 pending 成员，由业务同步及确认流程完成收尾；不释放执行租约或容量。
	CancelQueuedJob(ctx context.Context, sessionID string) (bool, error)
	// AcknowledgeJobDispatch 校验调度锁 token，确认 Job 为 dispatched、failed 或 cancelled 后，
	// 从待调度集合移除 SessionID，并将调度记录保留 24 小时；重复确认可安全重试。
	// 锁失效、记录不存在或状态不符返回 false, nil；不释放执行租约、容量或调度锁。
	AcknowledgeJobDispatch(ctx context.Context, sessionID, claimToken string) (bool, error)
	// RenewJobClaim 校验独立调度锁的 token 后按 ttl 续期，锁失效时返回 false, nil。
	RenewJobClaim(ctx context.Context, sessionID, token string, ttl time.Duration) (bool, error)
	// ReleaseJobClaim 仅删除 token 匹配的调度锁，锁不存在或已换持有者时不删除。
	ReleaseJobClaim(ctx context.Context, sessionID, token string) error
	// FailJob 在同一事务中校验领取 token 和有效期，将 Job 及对应 Session 标记为失败并保存原因。
	// 同时清除领取信息，使 Job 不再参与调度；不得覆盖已结束的 Session 状态。
	// 成功返回 true, nil；领取失效或 Job 已结束调度时返回 false, nil；存储异常返回错误。
	FailJob(ctx context.Context, sessionID, claimToken, reason string) (bool, error)
	// GetExecutionLease 查询 Session 当前的执行租约，不存在时返回 nil, nil。
	GetExecutionLease(ctx context.Context, sessionID string) (*ExecutionLease, error)
	// ReleaseExecutionLease 校验执行 token 后，原子释放对应租约和 Runner 容量占用。
	// 租约不存在或 token 不匹配时返回 false, nil，不影响其他执行。
	ReleaseExecutionLease(ctx context.Context, sessionID, executionToken string) (bool, error)
	// ReserveExecutionLease 为指定 Session 创建执行租约，并原子占用目标 Runner 的一个并发名额。
	// 检查 Runner 在线且未维护、Session 无已有租约；清理过期容量成员后检查剩余名额。
	// capacity 是调用方读取的并发上限，预占时须与最新心跳中的上限取较小值；ttl 必须为正数。
	// 成功时生成执行 token、设置租约期限并返回租约；已有租约、节点不可用或满载时返回错误。
	// 不更新 Job、Session 或调度状态，不投递消息；调度锁和固定 Job 的保存由外层负责。
	ReserveExecutionLease(ctx context.Context, taskID, sessionID, runnerID string, capacity int, ttl time.Duration) (*ExecutionLease, error)
	// ReserveExecution 原子校验调度领取 token 和有效期，并为 Job 预留目标 Runner 的并发名额。
	// 须检查 Runner 在线、支持 Job 的引擎且仍有容量；已绑定 Runner 的任务不得改投其他节点。
	// 首次预留生成执行 token，以存储端当前时间加 ttl 设置执行租约期限，ttl 必须为正数。
	// 同时保存 Job 和 Session 的 Runner 绑定，将 Job 状态设为 dispatching，返回更新后的 Job。
	// 重复预留须复用已保存且有效的同一执行租约，不能重复占用名额或更换目标 Runner。
	// 已有执行租约失效时返回错误，由调用方处理恢复，不得静默生成新租约重新投递。
	// 领取失效、Runner 不可用、容量不足或存储异常均返回错误；本方法不投递消息或启动执行。
	ReserveExecution(ctx context.Context, sessionID, claimToken, runnerID string, ttl time.Duration) (*Job, error)
	// ListRunners 返回执行节点的状态、引擎能力和容量快照，无节点时返回空列表。
	// Service 用于查找已绑定节点或选择候选节点，实际预留仍须原子复查节点状态与容量。
	ListRunners(ctx context.Context) ([]Runner, error)
	// PublishJob 将已确定的执行消息幂等投递至 job.RunnerID 对应的执行通道。
	// 未发布时按 Redis 时间校验消息携带的租约到期时间；使用固定的 MessageID 和消息内容重试。
	// 发布去重标记与消息写入须原子完成；不向 Runner 暴露调度领取 token 等内部字段。
	// 成功仅表示消息已投递，不表示 Runner 已开始执行；结果不确定时返回错误，供同消息重试。
	PublishJob(ctx context.Context, job *Job) error
	// PublishCommand 将控制指令发送至 command.RunnerID 对应的命令通道。
	// 同一命令重试时复用 ID 和目标执行 token，Runner 须支持重复消费并校验执行归属及有效期。
	// 成功仅表示命令已发送，不保证执行已停止；本方法不清除执行租约或释放容量。
	PublishCommand(ctx context.Context, command *RunnerCommand) error
}

// store 是统一存储入口，数据库连接及接口方法后续补充。
type store struct {
	redis *redis.Client
}

// NewStore 注入 Redis 客户端，客户端由调用方创建和关闭。
func NewStore(client *redis.Client) *store {
	return &store{redis: client}
}

const (
	jobKeyPrefix               = "infra-agent-cloud:{dispatch}:job:"
	jobQueueKey                = "infra-agent-cloud:{dispatch}:pending"
	executionLeaseKeyPrefix    = "infra-agent-cloud:{dispatch}:lease:"
	runnerCapacityKeyPrefix    = "infra-agent-cloud:{dispatch}:capacity:"
	runnerHeartbeatKeyPrefix   = "infra-agent-cloud:{dispatch}:heartbeat:"
	runnerHeartbeatIndexKey    = "infra-agent-cloud:{dispatch}:heartbeat:index"
	runnerMaintenanceKeyPrefix = "infra-agent-cloud:{dispatch}:maintenance:"
	runnerJobsKeyPrefix        = "infra-agent-cloud:{dispatch}:jobs:"
	jobPublishedKeyPrefix      = "infra-agent-cloud:{dispatch}:published:"
	runnerJobsMaxLen           = 10000
	jobRetentionSeconds        = 24 * 60 * 60
)

var enqueueJobScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
redis.call('SET', KEYS[1], ARGV[1])
redis.call('ZADD', KEYS[2], ARGV[2], ARGV[3])
return 1
`)

var saveJobScript = redis.NewScript(`
-- 校验当前调度锁，阻止已失去处理权的调度器覆盖记录。
if redis.call('GET', KEYS[2]) ~= ARGV[1] then return 0 end
-- 保存调用方准备好的状态和固定消息，保留待调度集合成员及当前锁。
redis.call('SET', KEYS[1], ARGV[2])
return 1
`)

var cancelQueuedJobScript = redis.NewScript(`
-- 仅在排队阶段取消，不能覆盖已经形成固定消息的投递阶段。
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
local job = cjson.decode(raw)
if job.Status == 'cancelled' then return 1 end
if job.Status ~= 'queued' then return 0 end
job.Status = 'cancelled'
redis.call('SET', KEYS[1], cjson.encode(job))
-- 使旧调度器的带 token 保存失效，pending 成员仍保留供后续收尾。
redis.call('DEL', KEYS[2])
return 1
`)

var acknowledgeJobDispatchScript = redis.NewScript(`
-- KEYS：1 为调度记录，2 为调度锁，3 为待调度集合。
-- ARGV：1 为锁 token，2 为 SessionID，3 为调度记录保留秒数。
if redis.call('GET', KEYS[2]) ~= ARGV[1] then return 0 end
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
-- 读取存储中的最新状态，不能移除仍在排队或投递的 Job。
local job = cjson.decode(raw)
if job.Status ~= 'dispatched' and job.Status ~= 'failed' and job.Status ~= 'cancelled' then return 0 end
-- 只结束调度；完整记录短期保留，执行租约、容量和调度锁由各自流程处理。
redis.call('ZREM', KEYS[3], ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`)

var renewJobClaimScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then return 0 end
return redis.call('PEXPIRE', KEYS[1], ARGV[2])
`)

var releaseJobClaimScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then return 0 end
return redis.call('DEL', KEYS[1])
`)

var releaseExecutionLeaseScript = redis.NewScript(`
local raw = redis.call('GET', KEYS[1])
if not raw then return 0 end
local lease = cjson.decode(raw)
if lease.ExecutionToken ~= ARGV[1] or lease.RunnerID ~= ARGV[2] then return 0 end
redis.call('ZREM', KEYS[2], ARGV[1])
redis.call('DEL', KEYS[1])
local latest = redis.call('ZREVRANGE', KEYS[2], 0, 0, 'WITHSCORES')
if #latest > 0 then redis.call('PEXPIREAT', KEYS[2], latest[2]) end
return 1
`)

var reserveExecutionLeaseScript = redis.NewScript(`
-- KEYS：1 为 Session 租约，2 为 Runner 容量集合，3 为维护标记，4 为心跳记录。
-- ARGV：1 为租约 JSON，2 为执行 token，3 为当前毫秒时间戳，
-- 4 为调用方容量上限，5 为租约到期毫秒时间戳，6 为租约有效期毫秒数。
-- 返回值：1 为成功，0 为满载，-1 为已有租约，-2 为节点不可用。

-- 维护中的节点不接收新任务。
if redis.call('EXISTS', KEYS[3]) == 1 then return -2 end
-- 心跳 key 不存在或节点非在线时拒绝预占；心跳 key 依靠自身 TTL 过期。
local raw = redis.call('GET', KEYS[4])
if not raw then return -2 end
local heartbeat = cjson.decode(raw)
if heartbeat.Status ~= 'online' then return -2 end
-- 同一 Session 不能覆盖已有租约，也不能重复占用名额。
if redis.call('EXISTS', KEYS[1]) == 1 then return -1 end
-- 容量成员的 score 是租约到期时间，先移除已经到期的占用。
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', ARGV[3])
-- 取调用方与最新心跳容量的较小值，防止使用旧快照超额分配。
local limit = math.min(tonumber(ARGV[4]), tonumber(heartbeat.Capacity or 0))
if redis.call('ZCARD', KEYS[2]) >= limit then return 0 end
-- 保存带有效期的租约，再以执行 token 为成员占用一个名额。
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[6])
redis.call('ZADD', KEYS[2], ARGV[5], ARGV[2])
-- 容量集合在剩余租约最晚到期时自动清理，避免提前删除仍有效的占用。
local latest = redis.call('ZREVRANGE', KEYS[2], 0, 0, 'WITHSCORES')
redis.call('PEXPIREAT', KEYS[2], latest[2])
return 1
`)

var publishJobScript = redis.NewScript(`
-- KEYS：1 为目标 Runner 的 Job Stream，2 为本次 Session 的发布标记。
-- ARGV：1 为固定消息 ID，2 为租约到期毫秒时间戳，3 为执行消息 JSON，4 为 Stream 长度上限。
-- 已发布的同一消息直接成功，不同消息返回冲突，避免重试重复写入 Stream。
local published = redis.call('GET', KEYS[2])
if published then
  if published == ARGV[1] then return 0 else return -2 end
end
-- 使用 Redis 当前时间判断消息携带的租约是否到期，不读取执行租约 key。
local now = redis.call('TIME')
local nowMs = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
if tonumber(ARGV[2]) <= nowMs then return -1 end
-- 写入执行消息，发布标记至少保留 24 小时，并覆盖消息携带的剩余租约期限。
redis.call('XADD', KEYS[1], 'MAXLEN', '~', ARGV[4], '*', 'payload', ARGV[3])
redis.call('SET', KEYS[2], ARGV[1], 'PX', math.max(86400000, tonumber(ARGV[2]) - nowMs))
return 1
`)

// EnqueueJob 按 SessionID 原子去重，保存 Job 并按 QueuedAt 加入待调度集合。
func (s *store) EnqueueJob(ctx context.Context, job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	if err := enqueueJobScript.Run(ctx, s.redis,
		[]string{jobKeyPrefix + job.SessionID, jobQueueKey},
		string(data), job.QueuedAt.UnixMilli(), job.SessionID,
	).Err(); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}
	return nil
}

func (s *store) ListPendingJobIDs(ctx context.Context) ([]string, error) {
	return s.redis.ZRange(ctx, jobQueueKey, 0, -1).Result()
}

// ClaimJob 为指定 Job 获取独立调度锁，token 在内部生成。
func (s *store) ClaimJob(ctx context.Context, sessionID string, ttl time.Duration) (string, error) {
	token := rand.Text()
	key := jobKeyPrefix + sessionID + ":claim"
	ok, err := s.redis.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("claim job: %w", err)
	}
	if !ok {
		return "", nil
	}
	return token, nil
}

func (s *store) GetJob(ctx context.Context, sessionID string) (*Job, error) {
	return nil, errors.New("job lookup is not implemented")
}

// SaveJob 通过 Lua 校验调度锁 token，并写入完整 Job。
func (s *store) SaveJob(ctx context.Context, job *Job, claimToken string) (bool, error) {
	data, err := json.Marshal(job)
	if err != nil {
		return false, fmt.Errorf("marshal job: %w", err)
	}
	key := jobKeyPrefix + job.SessionID
	result, err := saveJobScript.Run(ctx, s.redis, []string{key, key + ":claim"}, claimToken, string(data)).Int64()
	if err != nil {
		return false, fmt.Errorf("save job: %w", err)
	}
	return result == 1, nil
}

// CancelQueuedJob 不依赖调用方持有调度锁，以状态条件更新和锁失效阻止继续准备。
func (s *store) CancelQueuedJob(ctx context.Context, sessionID string) (bool, error) {
	key := jobKeyPrefix + sessionID
	result, err := cancelQueuedJobScript.Run(ctx, s.redis, []string{key, key + ":claim"}).Int64()
	if err != nil {
		return false, fmt.Errorf("cancel queued job: %w", err)
	}
	return result == 1, nil
}

// AcknowledgeJobDispatch 通过 Lua 校验调度锁及最新状态，并完成队列移除和记录过期设置。
func (s *store) AcknowledgeJobDispatch(ctx context.Context, sessionID, claimToken string) (bool, error) {
	key := jobKeyPrefix + sessionID
	result, err := acknowledgeJobDispatchScript.Run(ctx, s.redis,
		[]string{key, key + ":claim", jobQueueKey},
		claimToken, sessionID, jobRetentionSeconds,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("acknowledge job dispatch: %w", err)
	}
	return result == 1, nil
}

func (s *store) RenewJobClaim(ctx context.Context, sessionID, token string, ttl time.Duration) (bool, error) {
	key := jobKeyPrefix + sessionID + ":claim"
	result, err := renewJobClaimScript.Run(ctx, s.redis, []string{key}, token, ttl.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("renew job claim: %w", err)
	}
	return result == 1, nil
}

func (s *store) ReleaseJobClaim(ctx context.Context, sessionID, token string) error {
	key := jobKeyPrefix + sessionID + ":claim"
	if err := releaseJobClaimScript.Run(ctx, s.redis, []string{key}, token).Err(); err != nil {
		return fmt.Errorf("release job claim: %w", err)
	}
	return nil
}

func (s *store) GetExecutionLease(ctx context.Context, sessionID string) (*ExecutionLease, error) {
	data, err := s.redis.Get(ctx, executionLeaseKeyPrefix+sessionID).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get execution lease: %w", err)
	}
	var lease ExecutionLease
	if err := json.Unmarshal(data, &lease); err != nil {
		return nil, fmt.Errorf("decode execution lease: %w", err)
	}
	return &lease, nil
}

func (s *store) ReleaseExecutionLease(ctx context.Context, sessionID, executionToken string) (bool, error) {
	lease, err := s.GetExecutionLease(ctx, sessionID)
	if err != nil {
		return false, err
	}
	if lease == nil || lease.ExecutionToken != executionToken {
		return false, nil
	}
	result, err := releaseExecutionLeaseScript.Run(ctx, s.redis,
		[]string{executionLeaseKeyPrefix + sessionID, runnerCapacityKeyPrefix + lease.RunnerID},
		executionToken, lease.RunnerID,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("release execution lease: %w", err)
	}
	return result == 1, nil
}

// ReserveExecutionLease 通过 Lua 完成节点检查、执行租约创建和容量预占。
func (s *store) ReserveExecutionLease(ctx context.Context, taskID, sessionID, runnerID string, capacity int, ttl time.Duration) (*ExecutionLease, error) {
	if ttl.Milliseconds() <= 0 {
		return nil, errors.New("execution lease TTL must be at least one millisecond")
	}
	now := time.Now()
	lease := &ExecutionLease{
		TaskID:             taskID,
		SessionID:          sessionID,
		RunnerID:           runnerID,
		ExecutionToken:     rand.Text(),
		ExecutionExpiresAt: now.Add(ttl),
	}
	data, err := json.Marshal(lease)
	if err != nil {
		return nil, fmt.Errorf("marshal execution lease: %w", err)
	}
	result, err := reserveExecutionLeaseScript.Run(ctx, s.redis,
		[]string{
			executionLeaseKeyPrefix + sessionID,
			runnerCapacityKeyPrefix + runnerID,
			runnerMaintenanceKeyPrefix + runnerID,
			runnerHeartbeatKeyPrefix + runnerID,
		},
		string(data), lease.ExecutionToken, now.UnixMilli(), capacity,
		lease.ExecutionExpiresAt.UnixMilli(), ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return nil, fmt.Errorf("reserve execution lease: %w", err)
	}
	switch result {
	case 1:
		return lease, nil
	case -1:
		return nil, errors.New("session already has an active execution lease")
	case -2:
		return nil, fmt.Errorf("runner unavailable: %s", runnerID)
	default:
		return nil, fmt.Errorf("runner capacity unavailable: %s", runnerID)
	}
}

// ListRunners 返回心跳有效的执行节点快照，无节点时返回空列表。
// 在线状态和引擎能力由 Service 筛选，实际剩余容量在预占时原子确认。
func (s *store) ListRunners(ctx context.Context) ([]Runner, error) {
	now := time.Now()
	score := strconv.FormatInt(now.UnixMilli(), 10)
	if err := s.redis.ZRemRangeByScore(ctx, runnerHeartbeatIndexKey, "-inf", score).Err(); err != nil {
		return nil, fmt.Errorf("clean expired runner heartbeats: %w", err)
	}
	ids, err := s.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     runnerHeartbeatIndexKey,
		Start:   score,
		Stop:    "+inf",
		ByScore: true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("list runner heartbeat IDs: %w", err)
	}
	if len(ids) == 0 {
		return []Runner{}, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = runnerHeartbeatKeyPrefix + id
	}
	values, err := s.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("get runner heartbeats: %w", err)
	}
	runners := make([]Runner, 0, len(values))
	for i, value := range values {
		if value == nil {
			continue
		}
		var heartbeat Runner
		if err := json.Unmarshal([]byte(value.(string)), &heartbeat); err != nil {
			return nil, fmt.Errorf("decode runner heartbeat %q: %w", ids[i], err)
		}
		if !heartbeat.ExpiresAt.After(now) {
			continue
		}
		runners = append(runners, heartbeat)
	}
	return runners, nil
}

// PublishJob 幂等写入指定 Runner 的执行通道，不发送调度锁等 Server 内部控制字段。
func (s *store) PublishJob(ctx context.Context, job *Job) error {
	data, err := json.Marshal(map[string]any{
		"MessageID":          job.MessageID,
		"TaskID":             job.TaskID,
		"SessionID":          job.SessionID,
		"RunnerID":           job.RunnerID,
		"ExecutorSessionID":  job.ExecutorSessionID,
		"Prompt":             job.Prompt,
		"AgentVersionID":     job.AgentVersionID,
		"Engine":             job.Engine,
		"Model":              job.Model,
		"SystemPrompt":       job.SystemPrompt,
		"ExecutionToken":     job.ExecutionToken,
		"ExecutionExpiresAt": job.ExecutionExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("marshal execution message: %w", err)
	}
	result, err := publishJobScript.Run(ctx, s.redis,
		[]string{runnerJobsKeyPrefix + job.RunnerID, jobPublishedKeyPrefix + job.RunnerID + ":" + job.SessionID},
		job.MessageID, job.ExecutionExpiresAt.UnixMilli(), string(data), runnerJobsMaxLen,
	).Int64()
	if err != nil {
		return fmt.Errorf("publish job: %w", err)
	}
	switch result {
	case 0, 1:
		return nil
	case -1:
		return ErrExecutionLeaseExpired
	case -2:
		return ErrDispatchConflict
	default:
		return fmt.Errorf("unexpected job publish result: %d", result)
	}
}
