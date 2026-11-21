package watch

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NewHandler 创建发送全量快照、增量事件和心跳的 Watch 处理函数。
func NewHandler(store Store, broadcaster *EventBroadcaster) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		clusterID := c.Param("clusterID")
		events, disconnected, unsubscribe := broadcaster.Subscribe(clusterID)
		defer unsubscribe()

		var fullSync Event
		if store != nil {
			snapshot, err := store.GetSnapshot(ctx, clusterID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load watch snapshot"})
				return
			}
			fullSync, err = newFullSyncEvent(snapshot)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode watch snapshot"})
				return
			}
		}

		c.Header("Content-Type", "application/x-ndjson")
		c.Status(http.StatusOK)
		response := http.NewResponseController(c.Writer)
		if err := writeEvent(c, response, fullSync); err != nil {
			return
		}

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			var event Event
			select {
			case <-ctx.Done():
				return
			case <-disconnected:
				return
			case <-ticker.C:
				event = Event{Type: "heartbeat"}
			case event = <-events:
			}
			if err := writeEvent(c, response, event); err != nil {
				return
			}
		}
	}
}

// writeEvent 写入并刷新响应，对支持写入期限的连接限制阻塞时间。
func writeEvent(c *gin.Context, response *http.ResponseController, event Event) error {
	if err := response.SetWriteDeadline(time.Now().Add(5 * time.Second)); err == nil {
		defer response.SetWriteDeadline(time.Time{})
	}
	if event.Type != "" {
		if err := json.NewEncoder(c.Writer).Encode(event); err != nil {
			return err
		}
	}
	return response.Flush()
}

// newFullSyncEvent 将快照转换为 Watch 全量事件。
func newFullSyncEvent(snapshot Snapshot) (Event, error) {
	data, err := json.Marshal(map[string]any{
		"management": map[string]map[string]string{
			"eks_nodes":     snapshot.EKSNodes,
			"cluster_nodes": snapshot.ClusterNodes,
		},
		"clusterCRs": map[string][]json.RawMessage{"cluster_crs": snapshot.ClusterCRs},
	})
	if err != nil {
		return Event{}, err
	}
	return Event{
		Type: "fullSync",
		Data: string(data),
		Offset: map[string]uint64{
			"managementOffset": snapshot.ManagementOffset,
			"clusterCROffset":  snapshot.ClusterCROffset,
		},
	}, nil
}
