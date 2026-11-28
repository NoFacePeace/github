package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"infra-agent-cloud/protocol"
)

// Redis key 常量。
const (
	// Runner 的心跳记录和在线索引与 Server 调度读取使用同一组 key。
	runnerHeartbeatKeyPrefix = "infra-agent-cloud:{dispatch}:heartbeat:"
	runnerHeartbeatsKey      = "infra-agent-cloud:{dispatch}:heartbeats"
)

// PublishHeartbeat 上报 Runner 的心跳并刷新在线有效期。
//
//   - 心跳记录：infra-agent-cloud:{dispatch}:heartbeat:runner-1
//   - 在线索引：infra-agent-cloud:{dispatch}:heartbeats
func (r *Redis) PublishHeartbeat(ctx context.Context, heartbeat protocol.Heartbeat, ttl time.Duration) error {
	const publishHeartbeatScript = `
local heartbeatKey = KEYS[1]
local indexKey = KEYS[2]
local instanceId = ARGV[1]
local payload = ARGV[2]
local ttlMs = ARGV[3]
local expiresAtMs = ARGV[4]
local runnerId = ARGV[5]

-- 同一个 Runner ID 在上一实例的心跳过期前只能由该实例刷新。
local current = redis.call('GET', heartbeatKey)
if current then
  local ok, heartbeat = pcall(cjson.decode, current)
  if not ok or type(heartbeat) ~= 'table' or heartbeat.InstanceID ~= instanceId then
    return 0
  end
end

redis.call('SET', heartbeatKey, payload, 'PX', ttlMs)
redis.call('ZADD', indexKey, expiresAtMs, runnerId)
return 1
`

	heartbeatKey := runnerHeartbeatKeyPrefix + heartbeat.RunnerID
	if heartbeat.Status == "" {
		heartbeat.Status = "online"
	}
	heartbeat.ExpiresAt = time.Now().Add(ttl)
	payload, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("marshal runner heartbeat: %w", err)
	}
	result, err := r.client.Do(ctx, "EVAL", publishHeartbeatScript, 2,
		heartbeatKey, runnerHeartbeatsKey,
		heartbeat.InstanceID, string(payload), ttl.Milliseconds(),
		heartbeat.ExpiresAt.UnixMilli(), heartbeat.RunnerID).Int64()
	if err != nil {
		return fmt.Errorf("publish runner heartbeat: %w", err)
	}
	if result == 0 {
		return fmt.Errorf("runner heartbeat is owned by another instance: %s", heartbeat.RunnerID)
	}
	return nil
}
