package store

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"infra-agent-cloud/protocol"

	"github.com/redis/go-redis/v9"
)

// Redis key 常量。
const (
	// Runner 专属 Job Stream 的 key 前缀，完整示例为 infra-agent-cloud:runner:jobs:runner-1。
	runnerJobsKeyPrefix = "infra-agent-cloud:runner:jobs:"
	// Runner 向 Server 回传执行事件的共享 Stream。
	runnerEventsKey = "infra-agent-cloud:runner:events"
)

// Redis 消费者组常量。
const (
	// Job 消费者组名称的前缀，完整示例为 runner-jobs-group:runner-1。
	runnerJobsGroupPrefix = "runner-jobs-group:"
)

const runnerJobsMaxLen = 10000

var (
	// ErrLeaseExpired 表示尚未投递的 Job 携带的租约已过期。
	ErrLeaseExpired = errors.New("job lease expired")
	// ErrDispatchConflict 表示目标 Stream 上已为同一 Session 投递过另一条消息。
	ErrDispatchConflict = errors.New("job dispatch conflict")
)

// Redis 封装 Runner 执行链路的 Redis 操作。待实现的方法如下。
//
//   - ConsumeCommands：消费发给当前 Runner 的控制命令。
//   - PublishHeartbeat：刷新当前 Runner 的心跳和在线索引。
//   - RemoveHeartbeat：清理当前 Runner 的心跳和在线索引。
//   - GetMaintenance：读取当前 Runner 的维护状态。
//   - GetWorkspaceCleanupResult：查询工作区清理命令的结果。
//   - CompleteWorkspaceCleanup：保存工作区清理命令的结果。
type Redis struct {
	client       *redis.Client
	jobStreamKey string
	jobGroup     string
	jobConsumer  string
}

func NewRedis(client *redis.Client, runnerID string) Redis {
	return Redis{
		client:       client,
		jobStreamKey: runnerJobsKeyPrefix + url.PathEscape(runnerID),
		jobGroup:     runnerJobsGroupPrefix + runnerID,
		jobConsumer:  runnerID + "-" + rand.Text(),
	}
}

// ensureGroup 确保指定 Stream 的消费者组存在。
//
//   - Stream key 由调用方传入，Job Stream 示例：infra-agent-cloud:runner:jobs:runner-1
func (r *Redis) ensureGroup(ctx context.Context, stream, group string) error {
	err := r.client.Do(ctx, "XGROUP", "CREATE", stream, group, "0", "MKSTREAM").Err()
	if redis.HasErrorPrefix(err, "BUSYGROUP ") {
		return nil
	}
	return err
}

// PublishJob 幂等地将执行消息投递到目标 Runner 的 Job Stream。
//
//   - Job Stream：infra-agent-cloud:runner:jobs:runner-1
//   - 投递去重标记：{infra-agent-cloud:runner:jobs:runner-1}:published:session-1
//
// 去重标记以完整的 Stream key 作为 hash tag，确保两个 key 同槽。
func (r *Redis) PublishJob(ctx context.Context, job *protocol.Job) error {
	const publishJobScript = `
local streamKey = KEYS[1]
local publishedKey = KEYS[2]
local messageId = ARGV[1]
local leaseExpiresAtMs = tonumber(ARGV[2])
local payload = ARGV[3]
local maxStreamLength = ARGV[4]
local minMarkerTtlMs = 24 * 60 * 60 * 1000

-- 先查去重，让已经投递的消息在租约过期后重试时仍返回成功。
local publishedId = redis.call('GET', publishedKey)
if publishedId == messageId then
  return 0
end
if publishedId then
  return -2
end

-- 首次写入前按 Redis 自己的时钟检查租约。
local now = redis.call('TIME')
local nowMs = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
local remainingMs = leaseExpiresAtMs - nowMs
if remainingMs <= 0 then
  return -1
end

-- 去重标记至少保留 24 小时，同时覆盖租约剩余时间。
redis.call('XADD', streamKey, 'MAXLEN', '~', maxStreamLength, '*', 'payload', payload)
redis.call('SET', publishedKey, messageId, 'PX', math.max(minMarkerTtlMs, remainingMs))
return 1
`

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal execution message: %w", err)
	}
	stream := runnerJobsKeyPrefix + url.PathEscape(job.RunnerID)
	published := "{" + stream + "}:published:" + job.SessionID
	result, err := r.client.Do(ctx, "EVAL", publishJobScript, 2, stream, published,
		job.MessageID, job.LeaseExpiresAt.UnixMilli(), string(payload), runnerJobsMaxLen).Int64()
	if err != nil {
		return fmt.Errorf("publish job: %w", err)
	}
	switch result {
	case 0, 1:
		return nil
	case -1:
		return ErrLeaseExpired
	case -2:
		return ErrDispatchConflict
	default:
		return fmt.Errorf("unexpected job publish result: %d", result)
	}
}

// ClaimJob 从当前 Runner 的执行通道领取任务并返回 Redis 消息 ID。
//
//   - Job Stream：infra-agent-cloud:runner:jobs:runner-1
func (r *Redis) ClaimJob(ctx context.Context) (*protocol.Job, string, error) {
	claimArgs := []any{"XAUTOCLAIM", r.jobStreamKey, r.jobGroup, r.jobConsumer, 30000, "0-0", "COUNT", 1}
	reply, err := r.client.Do(ctx, claimArgs...).Result()
	if redis.HasErrorPrefix(err, "NOGROUP ") {
		if err := r.ensureGroup(ctx, r.jobStreamKey, r.jobGroup); err != nil {
			return nil, "", fmt.Errorf("create job consumer group: %w", err)
		}
		reply, err = r.client.Do(ctx, claimArgs...).Result()
	}
	if err != nil {
		return nil, "", fmt.Errorf("claim pending job: %w", err)
	}
	claim, ok := reply.([]any)
	if !ok || len(claim) < 2 {
		return nil, "", errors.New("unexpected XAUTOCLAIM reply")
	}
	messages, ok := claim[1].([]any)
	if !ok {
		return nil, "", errors.New("unexpected XAUTOCLAIM messages")
	}
	if len(messages) == 0 {
		// 客户端的读取超时要长于 Redis 的阻塞等待时间。
		reply, err = r.client.WithTimeout(15*time.Second).Do(ctx, "XREADGROUP", "GROUP", r.jobGroup, r.jobConsumer,
			"COUNT", 1, "BLOCK", 5000, "STREAMS", r.jobStreamKey, ">").Result()
		if errors.Is(err, redis.Nil) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", fmt.Errorf("read new job: %w", err)
		}
		// RESP2 返回 [stream, entries] 列表，RESP3 返回 stream 到 entries 的映射。
		switch streams := reply.(type) {
		case []any:
			if len(streams) == 0 {
				return nil, "", nil
			}
			stream, ok := streams[0].([]any)
			if !ok || len(stream) != 2 {
				return nil, "", errors.New("unexpected XREADGROUP stream")
			}
			messages, ok = stream[1].([]any)
			if !ok {
				return nil, "", errors.New("unexpected XREADGROUP messages")
			}
		case map[any]any:
			messages, ok = streams[r.jobStreamKey].([]any)
			if !ok {
				return nil, "", errors.New("unexpected XREADGROUP messages")
			}
		default:
			return nil, "", errors.New("unexpected XREADGROUP reply")
		}
	}

	if len(messages) == 0 {
		return nil, "", nil
	}
	entry, ok := messages[0].([]any)
	if !ok || len(entry) != 2 {
		return nil, "", errors.New("unexpected job stream entry")
	}
	entryID, ok := entry[0].(string)
	if !ok {
		return nil, "", errors.New("unexpected job stream entry ID")
	}
	fields, ok := entry[1].([]any)
	if !ok || len(fields)%2 != 0 {
		return nil, "", fmt.Errorf("invalid fields in job stream entry %q", entryID)
	}
	for i := 0; i < len(fields); i += 2 {
		if fields[i] != "payload" {
			continue
		}
		payload, ok := fields[i+1].(string)
		if !ok {
			break
		}
		var job protocol.Job
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			return nil, "", fmt.Errorf("decode job stream entry %q: %w", entryID, err)
		}
		return &job, entryID, nil
	}
	return nil, "", fmt.Errorf("job stream entry %q is missing a string payload", entryID)
}

// AckJob 确认当前 Runner 的 Job Stream 消息。
//
//   - Job Stream：infra-agent-cloud:runner:jobs:runner-1
func (r *Redis) AckJob(ctx context.Context, entryID string) error {
	return r.client.Do(ctx, "XACK", r.jobStreamKey, r.jobGroup, entryID).Err()
}

// PublishEvent 将 Runner 的执行事件写入 Redis 事件流。
//
//   - Event Stream：infra-agent-cloud:runner:events
func (r *Redis) PublishEvent(ctx context.Context, event *protocol.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal runner event: %w", err)
	}
	return r.client.Do(ctx, "XADD", runnerEventsKey, "*", "payload", string(payload)).Err()
}
