package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"infra-agent-cloud/protocol"
)

// Redis key 常量。
const (
	// Runner 向 Server 回传执行事件的共享 Stream。
	runnerEventsKey = "infra-agent-cloud:runner:events"
)

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
