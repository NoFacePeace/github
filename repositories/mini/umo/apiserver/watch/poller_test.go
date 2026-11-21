package watch

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

type pollStore struct {
	Store
	listManagement func(uint64) (ManagementDelta, error)
	listClusterCRs func(string, uint64) (ClusterCRDelta, error)
}

func (s pollStore) ListManagementSince(_ context.Context, offset uint64) (ManagementDelta, error) {
	return s.listManagement(offset)
}

func (s pollStore) ListClusterCRsSince(_ context.Context, clusterID string, offset uint64) (ClusterCRDelta, error) {
	return s.listClusterCRs(clusterID, offset)
}

// 验证定时轮询在失败后继续，并在上下文取消后退出。
func TestPollerRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		broadcaster := NewEventBroadcaster()
		events, _, unsubscribe := broadcaster.Subscribe("cluster-a")
		defer unsubscribe()
		attempts := 0
		p := poller{broadcaster: broadcaster, store: pollStore{
			listManagement: func(uint64) (ManagementDelta, error) {
				attempts++
				if attempts == 1 {
					return ManagementDelta{}, errors.New("query failed")
				}
				return ManagementDelta{EKSNodes: map[string]string{"node-a": "offline"}, Offset: 10}, nil
			},
			listClusterCRs: func(string, uint64) (ClusterCRDelta, error) {
				return ClusterCRDelta{}, nil
			},
		}}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			p.Run(ctx)
		}()
		synctest.Wait()

		time.Sleep(10 * time.Second)
		synctest.Wait()
		select {
		case event := <-events:
			if event.Offset["managementOffset"] != 10 {
				t.Fatalf("unexpected management event: %+v", event)
			}
		default:
			t.Fatal("no management event after retry")
		}

		cancel()
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("poller did not exit after cancellation")
		}
	})
}

// 验证运维和各集群 CR 的偏移量独立推进，失败的查询保留原偏移量。
func TestPollOnce(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	a, _, unsubscribeA := broadcaster.Subscribe("cluster-a")
	defer unsubscribeA()
	b, _, unsubscribeB := broadcaster.Subscribe("cluster-b")
	defer unsubscribeB()

	queryErr := errors.New("query failed")
	round := 1
	p := poller{broadcaster: broadcaster, store: pollStore{
		listManagement: func(offset uint64) (ManagementDelta, error) {
			if offset != 0 {
				t.Fatalf("management queried from %d, want 0", offset)
			}
			if round == 1 {
				return ManagementDelta{}, queryErr
			}
			return ManagementDelta{EKSNodes: map[string]string{"node-a": "offline"}, Offset: 10}, nil
		},
		listClusterCRs: func(clusterID string, offset uint64) (ClusterCRDelta, error) {
			switch clusterID {
			case "cluster-a":
				if round == 1 && offset == 0 {
					return ClusterCRDelta{ClusterCRs: []json.RawMessage{json.RawMessage(`{}`)}, Offset: 20}, nil
				}
				if round == 2 && offset == 20 {
					return ClusterCRDelta{}, nil
				}
			case "cluster-b":
				if offset == 0 {
					return ClusterCRDelta{}, nil
				}
			}
			t.Fatalf("unexpected CR query: round %d, cluster %s, offset %d", round, clusterID, offset)
			return ClusterCRDelta{}, nil
		},
	}}
	nextEvent := func(events <-chan Event) Event {
		t.Helper()
		select {
		case event := <-events:
			return event
		default:
			t.Fatal("expected an event")
			return Event{}
		}
	}
	if err := p.pollOnce(t.Context()); !errors.Is(err, queryErr) {
		t.Fatalf("first poll returned %v, want %v", err, queryErr)
	}
	if event := nextEvent(a); event.Type != "clusterCR" || event.Offset["clusterCROffset"] != 20 || event.Data != `{"cluster_crs":[{}]}` {
		t.Fatalf("unexpected CR event: %+v", event)
	}
	round++
	if err := p.pollOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, events := range []<-chan Event{a, b} {
		if event := nextEvent(events); event.Type != "management" || event.Offset["managementOffset"] != 10 {
			t.Fatalf("unexpected management event: %+v", event)
		}
	}
	if p.managementOffset != 10 {
		t.Fatalf("management offset = %d, want 10", p.managementOffset)
	}
}
