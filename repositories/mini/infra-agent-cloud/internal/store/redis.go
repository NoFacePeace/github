package store

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"infra-agent-cloud/runner"

	"github.com/redis/go-redis/v9"
)

// Redis key 常量。
const (
	// Runner 专属 Job Stream 的 key 前缀，完整示例为 infra-agent-cloud:runner:jobs:runner-1。
	runnerJobsKeyPrefix = "infra-agent-cloud:runner:jobs:"
)

// Redis 消费者组常量。
const (
	// Job 消费者组名称的前缀，完整示例为 runner-jobs-group:runner-1。
	runnerJobsGroupPrefix = "runner-jobs-group:"
)

// Redis 封装 Runner 执行链路的 Redis 操作。待实现的方法如下。
//
//   - ConsumeCommands：消费发给当前 Runner 的控制命令。
//   - PublishHeartbeat：刷新当前 Runner 的心跳和在线索引。
//   - RemoveHeartbeat：清理当前 Runner 的心跳和在线索引。
//   - GetMaintenance：读取当前 Runner 的维护状态。
//   - PublishEvent：回传执行事件并更新事件流的积压字节数。
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
func (r *Redis) ensureGroup(ctx context.Context, stream, group string) error {
	err := r.client.Do(ctx, "XGROUP", "CREATE", stream, group, "0", "MKSTREAM").Err()
	if redis.HasErrorPrefix(err, "BUSYGROUP ") {
		return nil
	}
	return err
}

// ClaimJob 从当前 Runner 的执行通道领取任务并返回 Redis 消息 ID。
func (r *Redis) ClaimJob(ctx context.Context) (*runner.Job, string, error) {
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
		var job runner.Job
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			return nil, "", fmt.Errorf("decode job stream entry %q: %w", entryID, err)
		}
		return &job, entryID, nil
	}
	return nil, "", fmt.Errorf("job stream entry %q is missing a string payload", entryID)
}
