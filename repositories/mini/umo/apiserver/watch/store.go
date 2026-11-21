package watch

import (
	"context"
	"encoding/json"
)

// Store 提供 Watch 全量同步和增量查询所需的数据。
type Store interface {
	GetSnapshot(ctx context.Context, clusterID string) (Snapshot, error)
	ListManagementSince(ctx context.Context, offset uint64) (ManagementDelta, error)
	ListClusterCRsSince(ctx context.Context, clusterID string, offset uint64) (ClusterCRDelta, error)
}

// ManagementDelta 包含运维增量及本次查询后的偏移量。
type ManagementDelta struct {
	EKSNodes     map[string]string
	ClusterNodes map[string]string
	Offset       uint64
}

// ClusterCRDelta 包含指定集群的 CR 增量及本次查询后的偏移量。
type ClusterCRDelta struct {
	ClusterCRs []json.RawMessage
	Offset     uint64
}

// Snapshot 包含运维状态、指定集群的 CR 配置及各自的偏移量。
type Snapshot struct {
	EKSNodes         map[string]string
	ClusterNodes     map[string]string
	ClusterCRs       []json.RawMessage
	ManagementOffset uint64
	ClusterCROffset  uint64
}
