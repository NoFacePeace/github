package watch

import "sync"

// Event 表示一条 Watch 响应事件。
type Event struct {
	Type   string            `json:"eventType"`
	Data   string            `json:"data,omitempty"`
	Offset map[string]uint64 `json:"eventOffset,omitempty"`
}

type subscription struct {
	clusterID string
	events    chan Event
	done      chan struct{}
}

// EventBroadcaster 将增量事件投递给订阅连接。
type EventBroadcaster struct {
	mu   sync.Mutex
	subs map[*subscription]struct{}
}

// NewEventBroadcaster 创建一个内存广播器。
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{subs: make(map[*subscription]struct{})}
}

// Subscribe 注册连接，返回事件队列、断开信号和注销函数。
func (b *EventBroadcaster) Subscribe(clusterID string) (events <-chan Event, done <-chan struct{}, unsubscribe func()) {
	sub := &subscription{clusterID: clusterID, events: make(chan Event, 100), done: make(chan struct{})}
	b.mu.Lock()
	b.subs[sub] = struct{}{}
	b.mu.Unlock()

	return sub.events, sub.done, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.remove(sub)
	}
}

// Publish 广播运维事件，或将 CR 增量投递给指定集群。
func (b *EventBroadcaster) Publish(clusterID string, event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for sub := range b.subs {
		if event.Type != "management" && (event.Type != "clusterCR" || sub.clusterID != clusterID) {
			continue
		}
		select {
		case sub.events <- event:
		default:
			b.remove(sub)
		}
	}
}

// subscribedClusters 返回当前有订阅的集群集合。
func (b *EventBroadcaster) subscribedClusters() map[string]struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	clusters := make(map[string]struct{})
	for sub := range b.subs {
		clusters[sub.clusterID] = struct{}{}
	}
	return clusters
}

// remove 在持锁时注销连接并通知 Handler 退出。
func (b *EventBroadcaster) remove(sub *subscription) {
	if _, exists := b.subs[sub]; exists {
		delete(b.subs, sub)
		close(sub.done)
	}
}
