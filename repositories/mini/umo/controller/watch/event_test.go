package watch

import (
	"context"
	"testing"

	"umo/controller/healthcheck"
)

// 验证两种事件里的运维数据均按协议字段解码并提交。
func TestHandleEventSubmitsManagement(t *testing.T) {
	const management = `{"eks_nodes":{"node-a":"manual_offline"},"cluster_nodes":{"pod-a":"disk_read_only"}}`
	for _, tc := range []struct {
		event      Event
		wantSubmit bool
	}{
		{Event{Type: EventTypeFullSync, Data: `{"management":` + management + `}`}, true},
		{Event{Type: EventTypeManagement, Data: management}, true},
		{Event{Type: EventTypeFullSync, Data: `{"management":null}`}, false},
		{Event{Type: EventTypeHeartbeat}, false},
	} {
		var submitted []healthcheck.Event
		c := client{submitEvent: func(_ context.Context, event healthcheck.Event) error {
			submitted = append(submitted, event)
			return nil
		}}
		if err := c.handleEvent(t.Context(), tc.event); err != nil {
			t.Fatal(err)
		}
		if tc.wantSubmit {
			if len(submitted) != 1 || submitted[0].EKSNodes["node-a"] != healthcheck.StatusNodeManualOffline ||
				submitted[0].ClusterNodes["pod-a"] != healthcheck.StatusDiskReadOnly {
				t.Errorf("%s submitted %v, want node and pod statuses", tc.event.Type, submitted)
			}
		} else if len(submitted) != 0 {
			t.Errorf("%s submitted %v, want no events", tc.event.Type, submitted)
		}
	}
}
