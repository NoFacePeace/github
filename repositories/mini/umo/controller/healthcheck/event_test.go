package healthcheck

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"umo/controller/k8s"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// 验证 Watch 事件复用动作流程，同一 Pod 优先采用节点异常。
func TestSubmitEvent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		checkerCalls := 0
		checker := testChecker{name: "test", check: func(context.Context) (map[string]Status, error) {
			checkerCalls++
			return nil, nil
		}}
		created, err := NewManager([]Checker{checker}, Config{Policies: []Policy{
			{Action: ActionManual, Status: []Status{StatusDiskReadOnly}, Threshold: 1, ReconcileInterval: time.Minute},
			{Action: ActionMigrate, Status: []Status{StatusNodeManualOffline}, Threshold: 1, ReconcileInterval: time.Minute},
		}})
		if err != nil {
			t.Fatal(err)
		}
		m := created.(*manager)
		newPod := func(name, node string) *corev1.Pod {
			pod := &corev1.Pod{}
			pod.Name, pod.Namespace, pod.Spec.NodeName = name, "default", node
			return pod
		}
		node := &corev1.Node{}
		node.Name = "offline"
		m.k8sClient = fake.NewClientBuilder().WithObjects(
			node, newPod("affected", node.Name), newPod("direct", "online"), newPod("untouched", "online"),
		).Build()
		event := Event{
			EKSNodes:     map[string]Status{node.Name: StatusNodeManualOffline},
			ClusterNodes: map[string]Status{"affected": StatusDiskReadOnly, "direct": StatusDiskReadOnly},
		}
		if err := created.SubmitEvent(t.Context(), event); err != nil {
			t.Fatalf("submit before Run: %v", err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			defer close(done)
			created.Run(ctx)
		}()
		defer func() {
			cancel()
			<-done
		}()
		synctest.Wait()

		for name, want := range map[string][2]string{
			"affected":  {string(ActionMigrate), ""},
			"direct":    {"", "true"},
			"untouched": {"", ""},
		} {
			pod := &corev1.Pod{}
			if err := m.k8sClient.Get(t.Context(), client.ObjectKey{Namespace: "default", Name: name}, pod); err != nil {
				t.Fatal(err)
			}
			if got := [2]string{pod.Labels[k8s.LabelHealthStatus], pod.Labels[k8s.LabelManualManagement]}; got != want {
				t.Errorf("pod %q action labels = %v, want %v", name, got, want)
			}
		}
		if err := m.k8sClient.Get(t.Context(), client.ObjectKey{Name: node.Name}, node); err != nil {
			t.Fatal(err)
		}
		if len(node.Labels) != 0 || checkerCalls != 0 {
			t.Fatalf("node labels = %v, checker calls = %d", node.Labels, checkerCalls)
		}
	})
}
