package watch

import (
	"context"
	"encoding/json"

	"umo/controller/healthcheck"
)

// EventType 表示 Watch 流中的事件类型。
type EventType string

const (
	EventTypeFullSync   EventType = "fullSync"
	EventTypeManagement EventType = "management"
	EventTypeClusterCR  EventType = "clusterCR"
	EventTypeHeartbeat  EventType = "heartbeat"
)

// Event 表示从 Watch 响应流解码的事件。
type Event struct {
	Type   EventType   `json:"eventType"`
	Data   string      `json:"data"`
	Offset EventOffset `json:"eventOffset"`
}

// EventOffset 表示服务端随事件发送的运维进度。
type EventOffset struct {
	ManagementOffset uint64 `json:"managementOffset"`
}

type managementData struct {
	EKSNodes     map[string]healthcheck.Status `json:"eks_nodes"`
	ClusterNodes map[string]healthcheck.Status `json:"cluster_nodes"`
}

// handleEvent 将全量快照和增量事件中的运维数据提交给健康管理器。
func (c client) handleEvent(ctx context.Context, event Event) error {
	var management *managementData
	switch event.Type {
	case EventTypeFullSync:
		var fullSync struct {
			Management *managementData `json:"management"`
		}
		if err := json.Unmarshal([]byte(event.Data), &fullSync); err != nil {
			return err
		}
		management = fullSync.Management
	case EventTypeManagement:
		management = &managementData{}
		if err := json.Unmarshal([]byte(event.Data), management); err != nil {
			return err
		}
	}
	if management == nil {
		return nil
	}
	return c.submitEvent(ctx, healthcheck.Event{
		EKSNodes:     management.EKSNodes,
		ClusterNodes: management.ClusterNodes,
	})
}
