package healthcheck

import (
	"context"
	"maps"
)

const eventQueueCapacity = 100

// Event 表示外部上报的节点和 Pod 异常。
type Event struct {
	EKSNodes     map[string]Status // Kubernetes Node 名称对应的异常状态。
	ClusterNodes map[string]Status // 受管 Pod 名称对应的异常状态。
}

// SubmitEvent 将事件入队，队列满时等待空位或调用方取消。
func (m *manager) SubmitEvent(ctx context.Context, event Event) error {
	// 入队后调用方可继续使用原始映射。
	event = Event{EKSNodes: maps.Clone(event.EKSNodes), ClusterNodes: maps.Clone(event.ClusterNodes)}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.events <- event:
		return nil
	}
}

// processEvent 将节点异常映射到 Pod，同一 Pod 优先采用节点异常。
func (m *manager) processEvent(ctx context.Context, event Event) {
	pods := m.listPods(ctx)
	if pods == nil {
		return
	}
	statuses := make(map[string][]Status)
	for name, pod := range pods {
		if status, exists := event.EKSNodes[pod.Spec.NodeName]; exists {
			statuses[name] = []Status{status}
		} else if status, exists := event.ClusterNodes[name]; exists {
			statuses[name] = []Status{status}
		}
	}
	m.processPodStatuses(ctx, statuses, pods)
}
