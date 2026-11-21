package watch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

type poller struct {
	store            Store
	broadcaster      *EventBroadcaster
	managementOffset uint64
	clusterCROffsets map[string]uint64
}

// RunPoller 使用给定的数据源和广播器运行增量轮询。
func RunPoller(ctx context.Context, store Store, broadcaster *EventBroadcaster) {
	p := poller{store: store, broadcaster: broadcaster}
	p.Run(ctx)
}

// Run 定时轮询增量，直到上下文取消。
func (p *poller) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.pollOnce(ctx); err != nil {
				log.Printf("watch poll failed: %v", err)
			}
		}
	}
}

// pollOnce 查询并广播一轮增量，各类查询失败时保留对应的偏移量。
func (p *poller) pollOnce(ctx context.Context) error {
	clusters := p.broadcaster.subscribedClusters()
	if len(clusters) == 0 {
		return nil
	}

	var pollErr error
	management, err := p.store.ListManagementSince(ctx, p.managementOffset)
	if err != nil {
		pollErr = fmt.Errorf("list management events: %w", err)
	} else if management.Offset > p.managementOffset && (len(management.EKSNodes) > 0 || len(management.ClusterNodes) > 0) {
		data, err := json.Marshal(map[string]map[string]string{
			"eks_nodes":     management.EKSNodes,
			"cluster_nodes": management.ClusterNodes,
		})
		if err != nil {
			pollErr = fmt.Errorf("encode management events: %w", err)
		} else {
			p.broadcaster.Publish("", Event{
				Type:   "management",
				Data:   string(data),
				Offset: map[string]uint64{"managementOffset": management.Offset},
			})
			p.managementOffset = management.Offset
		}
	}

	if p.clusterCROffsets == nil {
		p.clusterCROffsets = make(map[string]uint64)
	}
	for clusterID := range clusters {
		offset := p.clusterCROffsets[clusterID]
		delta, err := p.store.ListClusterCRsSince(ctx, clusterID, offset)
		if err != nil {
			pollErr = errors.Join(pollErr, fmt.Errorf("list CR events for cluster %q: %w", clusterID, err))
			continue
		}
		if delta.Offset <= offset || len(delta.ClusterCRs) == 0 {
			continue
		}
		data, err := json.Marshal(map[string][]json.RawMessage{"cluster_crs": delta.ClusterCRs})
		if err != nil {
			pollErr = errors.Join(pollErr, fmt.Errorf("encode CR events for cluster %q: %w", clusterID, err))
			continue
		}
		p.broadcaster.Publish(clusterID, Event{
			Type:   "clusterCR",
			Data:   string(data),
			Offset: map[string]uint64{"clusterCROffset": delta.Offset},
		})
		p.clusterCROffsets[clusterID] = delta.Offset
	}
	return pollErr
}
