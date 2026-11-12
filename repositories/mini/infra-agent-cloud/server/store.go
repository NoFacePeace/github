package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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
	// ListRecoverableSessions 分页查询恢复候选：早于 startupBefore 创建且尚未启动执行的 running 会话、
	// 执行租约已失效的 running 会话，以及已 failed 或 stopped 但仍持有有效执行租约的会话。
	// 租约有效性以存储端当前时间判断；候选仅供 Service 复查，不代表可直接更新状态。
	// cursor 为空表示第一页，limit 须为正数；使用稳定游标分页，返回的下一页游标为空表示扫描结束。
	// Service 须结合 Job 调度状态和截止时间复查，不能仅因尚无执行租约就判定排队会话失败。
	// TODO: 补充 Session 创建时间字段，供长期未启动会话的筛选使用。
	ListRecoverableSessions(ctx context.Context, startupBefore time.Time, cursor string, limit int) ([]Session, string, error)
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
	// 须校验执行租约有效，使用固定的 MessageID 和消息内容；重试不得重复发布同一消息。
	// 发布去重标记与消息写入须原子完成；不向 Runner 暴露调度领取 token 等内部字段。
	// 成功仅表示消息已投递，不表示 Runner 已开始执行；结果不确定时返回错误，供同消息重试。
	PublishJob(ctx context.Context, job *Job) error
	// PublishCommand 将控制指令发送至 command.RunnerID 对应的命令通道。
	// 同一命令重试时复用 ID 和目标执行 token，Runner 须支持重复消费并校验执行归属及有效期。
	// 成功仅表示命令已发送，不保证执行已停止；本方法不清除执行租约或释放容量。
	PublishCommand(ctx context.Context, command *RunnerCommand) error
	// CompleteJobDispatch 原子校验调度领取 token 和有效期，将 Job 标记为 dispatched 并清除领取信息。
	// 保留执行 token 和执行租约期限，Session 保持 running，实际启动进度由 Runner 事件表达。
	// 同一有效领取下重复确认已投递记录应视为成功，不覆盖取消或失败等终态。
	// 成功返回 true, nil；领取失效或状态不允许确认时返回 false, nil；存储异常返回错误。
	CompleteJobDispatch(ctx context.Context, sessionID, claimToken string) (bool, error)
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
	jobKeyPrefix            = "infra-agent-cloud:{dispatch}:job:"
	jobQueueKey             = "infra-agent-cloud:{dispatch}:pending"
	executionLeaseKeyPrefix = "infra-agent-cloud:{dispatch}:lease:"
	runnerCapacityKeyPrefix = "infra-agent-cloud:{dispatch}:capacity:"
)

var enqueueJobScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
redis.call('SET', KEYS[1], ARGV[1])
redis.call('ZADD', KEYS[2], ARGV[2], ARGV[3])
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
