package healthcheck

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type testChecker struct {
	name  string
	check func(context.Context) (map[string]Status, error)
}

func (c testChecker) GetName() string { return c.name }

func (c testChecker) Check(ctx context.Context) (map[string]Status, error) {
	return c.check(ctx)
}

// 验证检查器并发启动，且整轮调用等待全部检查结束。
func TestRunOnceConcurrentCheckers(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	started := make(chan struct{}, 2)
	releaseFirst := make(chan struct{})
	checkers := []Checker{
		testChecker{name: "first", check: func(received context.Context) (map[string]Status, error) {
			if received != ctx {
				return nil, errors.New("unexpected context")
			}
			started <- struct{}{}
			select {
			case <-releaseFirst:
				return map[string]Status{
					"default/shared": StatusNodeNotReady,
					"default/first":  StatusDiskReadOnly,
				}, nil
			case <-received.Done():
				return nil, received.Err()
			}
		}},
		testChecker{name: "second", check: func(received context.Context) (map[string]Status, error) {
			if received != ctx {
				return nil, errors.New("unexpected context")
			}
			started <- struct{}{}
			return map[string]Status{
				"default/shared": StatusDiskNoSpace,
				"default/second": StatusNodeUnreachable,
			}, nil
		}},
	}
	created, err := NewManager(checkers, Config{})
	if err != nil {
		t.Fatal(err)
	}
	m := created.(*manager)
	m.k8sClient = fake.NewClientBuilder().Build()
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.runOnce(ctx)
	}()

	// 第一个检查器尚未结束时，第二个检查器也必须已经启动。
	for range len(checkers) {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("checkers did not start concurrently")
		}
	}
	select {
	case <-done:
		t.Fatal("runOnce returned before all checkers finished")
	default:
	}
	close(releaseFirst)

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("runOnce did not finish")
	}
}

// 验证单个检查器失败不会阻止其他检查器执行。
func TestRunOnceCheckerFailure(t *testing.T) {
	otherCalled := false
	m := &manager{k8sClient: fake.NewClientBuilder().Build(), checkers: []Checker{
		testChecker{name: "failed", check: func(context.Context) (map[string]Status, error) {
			return nil, errors.New("source unavailable")
		}},
		testChecker{name: "other", check: func(context.Context) (map[string]Status, error) {
			otherCalled = true
			return map[string]Status{"default/other": StatusNodeNotReady}, nil
		}},
	}}

	m.runOnce(t.Context())
	if !otherCalled {
		t.Fatal("other checker was not called")
	}
}

// 验证两张记录表分别按动作窗口过期，恰好处于窗口边界的记录仍保留。
func TestClearExpiredPods(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()
		boundary := now.Add(-10 * time.Minute)
		recent := now.Add(-time.Minute)
		m := &manager{
			config: Config{Policies: []Policy{
				{Action: ActionReCreate, ReconcileInterval: 10 * time.Minute},
				{Action: ActionManual, ReconcileInterval: 30 * time.Minute},
			}},
			podActionMatchedAt: map[string]map[Action]time.Time{
				"expired": {ActionReCreate: now.Add(-11 * time.Minute)},
				"mixed": {
					ActionReCreate: now.Add(-11 * time.Minute),
					ActionManual:   now.Add(-11 * time.Minute),
				},
				"boundary": {ActionReCreate: boundary},
				"empty":    {},
			},
			podActionExecutedAt: map[string]map[Action]time.Time{
				"expired": {ActionManual: now.Add(-31 * time.Minute)},
				"mixed":   {ActionReCreate: recent},
			},
		}

		m.clearExpiredPods()

		wantMatched := map[string]map[Action]time.Time{
			"mixed":    {ActionManual: now.Add(-11 * time.Minute)},
			"boundary": {ActionReCreate: boundary},
		}
		wantExecuted := map[string]map[Action]time.Time{
			"mixed": {ActionReCreate: recent},
		}
		if !reflect.DeepEqual(m.podActionMatchedAt, wantMatched) {
			t.Fatalf("matched records = %v, want %v", m.podActionMatchedAt, wantMatched)
		}
		if !reflect.DeepEqual(m.podActionExecutedAt, wantExecuted) {
			t.Fatalf("executed records = %v, want %v", m.podActionExecutedAt, wantExecuted)
		}
	})
}

// 验证阈值边界、动作隔离及统计范围，无 Pod 时暂停处理。
func TestExceedsThreshold(t *testing.T) {
	now := time.Now()
	m := &manager{podActionMatchedAt: map[string]map[Action]time.Time{
		"pod-a": {ActionReCreate: now},
		"pod-b": {ActionReCreate: now},
		"gone":  {ActionReCreate: now},
	}}
	for _, tt := range []struct {
		name      string
		podNames  []string
		action    Action
		threshold float64
		want      bool
	}{
		{"equal threshold", []string{"pod-a", "pod-b", "pod-c", "pod-d"}, ActionReCreate, 0.5, false},
		{"above threshold", []string{"pod-a", "pod-b", "pod-c", "pod-d"}, ActionReCreate, 0.25, true},
		{"different action", []string{"pod-a", "pod-b"}, ActionManual, 0, false},
		{"departed pods excluded", []string{"pod-c", "pod-d"}, ActionReCreate, 0, false},
		{"no pods", nil, ActionReCreate, 1, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pods := make(map[string]corev1.Pod)
			for _, name := range tt.podNames {
				pods[name] = corev1.Pod{}
			}
			policy := Policy{Action: tt.action, Threshold: tt.threshold}
			if got := m.exceedsThreshold(policy, pods); got != tt.want {
				t.Fatalf("exceedsThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 验证先清理、再记录全部匹配结果；超限不会丢失记录或刷新执行时间。
func TestProcessStatusesCumulativeMatches(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()
		previous := now.Add(-time.Minute)
		cfg := Config{Policies: []Policy{
			{Action: ActionReCreate, Status: []Status{StatusNodeNotReady}, ReconcileInterval: 10 * time.Minute, Threshold: 0.25},
			{Action: ActionManual, Status: []Status{StatusDiskReadOnly}, ReconcileInterval: 10 * time.Minute, Threshold: 1},
		}}
		created, err := NewManager(nil, cfg)
		if err != nil {
			t.Fatal(err)
		}
		m := created.(*manager)
		pods := &corev1.PodList{}
		podMap := make(map[string]corev1.Pod)
		for _, name := range []string{"pod-a", "pod-b", "pod-c", "pod-d"} {
			pod := corev1.Pod{}
			pod.Name = name
			pod.Namespace = "default"
			pods.Items = append(pods.Items, pod)
			podMap[name] = pod
		}
		m.k8sClient = fake.NewClientBuilder().WithLists(pods).Build()
		m.podActionMatchedAt["pod-a"] = map[Action]time.Time{ActionReCreate: previous}
		m.podActionMatchedAt["pod-d"] = map[Action]time.Time{ActionReCreate: now.Add(-11 * time.Minute)}
		m.podActionExecutedAt["pod-a"] = map[Action]time.Time{ActionReCreate: previous}

		m.processStatuses(t.Context(), map[string][]Status{
			"pod-a":   {StatusNodeNotReady, StatusDiskReadOnly},
			"pod-b":   {StatusNodeNotReady},
			"pod-c":   {StatusDiskReadOnly},
			"missing": {StatusNodeNotReady},
		})
		want := map[string]map[Action]time.Time{
			"pod-a": {ActionReCreate: now},
			"pod-b": {ActionReCreate: now},
			"pod-c": {ActionManual: now},
		}
		if !reflect.DeepEqual(m.podActionMatchedAt, want) {
			t.Fatalf("matched records = %v, want %v", m.podActionMatchedAt, want)
		}
		wantExecuted := map[string]map[Action]time.Time{
			"pod-a": {ActionReCreate: previous},
			"pod-c": {ActionManual: now},
		}
		if !reflect.DeepEqual(m.podActionExecutedAt, wantExecuted) {
			t.Fatalf("executed records changed: %v", m.podActionExecutedAt)
		}
		if !m.exceedsThreshold(cfg.Policies[0], podMap) || m.exceedsThreshold(cfg.Policies[1], podMap) {
			t.Fatal("actions did not evaluate their own matched ratios")
		}

		time.Sleep(time.Minute)
		m.processStatuses(t.Context(), map[string][]Status{"pod-b": {StatusNodeNotReady}})
		if !m.exceedsThreshold(cfg.Policies[0], podMap) {
			t.Fatal("previous matches within the window were not counted")
		}
	})
}

// 验证动作标签、冷却去重及失败后的重试，不修改其他标签或注解。
func TestProcessStatusesActions(t *testing.T) {
	for _, tt := range []struct {
		name      string
		action    Action
		manual    bool
		failFirst bool
		wantKey   string
		wantValue string
	}{
		{name: "migrate", action: ActionMigrate, wantKey: "health-status", wantValue: "Migrate"},
		{name: "recreate", action: ActionReCreate, wantKey: "health-status", wantValue: "ReCreate"},
		{name: "in-place update", action: ActionInPlaceUpdate, wantKey: "health-status", wantValue: "InPlaceUpdate"},
		{name: "manual", action: ActionManual, wantKey: "manual-management", wantValue: "true"},
		{name: "automatic", action: ActionAutomatic, manual: true, wantKey: "manual-management", wantValue: "false"},
		{name: "manual blocks recreate", action: ActionReCreate, manual: true},
		{name: "manual blocks migrate", action: ActionMigrate, manual: true},
		{name: "manual blocks in-place update", action: ActionInPlaceUpdate, manual: true},
		{name: "unknown action", action: "Unknown"},
		{name: "retry failed patch", action: ActionReCreate, failFirst: true, wantKey: "health-status", wantValue: "ReCreate"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				created, err := NewManager(nil, Config{Policies: []Policy{{
					Action: tt.action, Status: []Status{StatusNodeNotReady},
					ReconcileInterval: time.Minute, Threshold: 1,
				}}})
				if err != nil {
					t.Fatal(err)
				}
				m := created.(*manager)
				pod := &corev1.Pod{}
				pod.Name, pod.Namespace = "pod-a", "default"
				pod.Labels = map[string]string{"keep": "value"}
				pod.Annotations = map[string]string{"keep": "annotation"}
				if tt.manual {
					pod.Labels["manual-management"] = "true"
				}
				attempts := 0
				m.k8sClient = fake.NewClientBuilder().WithObjects(pod).WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
						attempts++
						if tt.failFirst && attempts == 1 {
							return errors.New("patch unavailable")
						}
						return c.Patch(ctx, obj, patch, opts...)
					},
				}).Build()
				statuses := map[string][]Status{"pod-a": {StatusNodeNotReady}}
				m.processStatuses(t.Context(), statuses)
				if tt.failFirst {
					if len(m.podActionExecutedAt) != 0 {
						t.Fatal("failed patch recorded as executed")
					}
					m.processStatuses(t.Context(), statuses)
				}
				if tt.wantKey == "" {
					if attempts != 0 || len(m.podActionExecutedAt) != 0 {
						t.Fatal("skipped action was executed")
					}
					return
				}
				got := &corev1.Pod{}
				if err := m.k8sClient.Get(t.Context(), client.ObjectKeyFromObject(pod), got); err != nil {
					t.Fatal(err)
				}
				if got.Labels[tt.wantKey] != tt.wantValue || got.Labels["keep"] != "value" || got.Annotations["keep"] != "annotation" {
					t.Fatalf("unexpected metadata: %v", got.ObjectMeta)
				}
				executedAt := m.podActionExecutedAt[pod.Name][tt.action]
				if !executedAt.Equal(time.Now()) {
					t.Fatalf("executed time = %v, want %v", executedAt, time.Now())
				}
				previousAttempts := attempts
				time.Sleep(time.Second)
				m.processStatuses(t.Context(), statuses)
				if attempts != previousAttempts || !m.podActionExecutedAt[pod.Name][tt.action].Equal(executedAt) {
					t.Fatal("cooldown did not prevent repeated execution")
				}
				time.Sleep(time.Minute)
				m.processStatuses(t.Context(), statuses)
				if attempts != previousAttempts+1 || !m.podActionExecutedAt[pod.Name][tt.action].Equal(time.Now()) {
					t.Fatal("action was not executed after cooldown expired")
				}
			})
		})
	}
}
