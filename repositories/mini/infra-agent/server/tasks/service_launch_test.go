package tasks

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
)

type launchStore struct {
	stubTaskStore
	acquire    func(context.Context, string, string, time.Duration) (*lease, bool, error)
	release    func(context.Context, string, string, int64) error
	transition func(context.Context, string, UpdateRunStatusRequest) (bool, error)
}

func (s *launchStore) UpdateRunStatus(ctx context.Context, id string, request UpdateRunStatusRequest) (bool, error) {
	if s.transition != nil {
		return s.transition(ctx, id, request)
	}
	return s.stubTaskStore.UpdateRunStatus(ctx, id, request)
}

type launchRunnerFunc func(context.Context, agentloop.Input, llm.Model, agentloop.ToolExecutor, agentloop.Recorder) error

func (f launchRunnerFunc) Run(ctx context.Context, input agentloop.Input, model llm.Model, executor agentloop.ToolExecutor, recorder agentloop.Recorder) error {
	return f(ctx, input, model, executor, recorder)
}

func (s *launchStore) AcquireLease(ctx context.Context, id, owner string, ttl time.Duration) (*lease, bool, error) {
	return s.acquire(ctx, id, owner, ttl)
}

func (s *launchStore) ReleaseLease(ctx context.Context, id, owner string, epoch int64) error {
	return s.release(ctx, id, owner, epoch)
}

func TestLaunchAcquiresAndReleasesLease(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		released := false
		store := &launchStore{
			acquire: func(got context.Context, id, owner string, ttl time.Duration) (*lease, bool, error) {
				if got != ctx || id != "run_1" || owner != "executor_1" || ttl != 40*time.Second {
					t.Error("unexpected lease acquisition arguments")
				}
				// 模拟请求在租约获取成功后结束，后台清理仍应可用。
				cancel()
				return &lease{Owner: owner, Epoch: 7}, true, nil
			},
			release: func(cleanupCtx context.Context, id, owner string, epoch int64) error {
				released = true
				if cleanupCtx.Err() != nil || id != "run_1" || owner != "executor_1" || epoch != 7 {
					t.Error("cleanup must use an uncancelled context and the acquired credential")
				}
				deadline, ok := cleanupCtx.Deadline()
				if !ok || time.Until(deadline) != 2*time.Second {
					t.Error("cleanup must have the configured timeout")
				}
				return nil
			},
		}
		s := &service{executions: &executionRegistry{}, store: store, runner: agentloop.New(agentloop.Config{}), executorID: "executor_1", config: Config{LeaseTTL: 40 * time.Second, CleanupTimeout: 2 * time.Second}}
		if err := s.launch(ctx, Task{}, run{ID: "run_1"}); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		if !released {
			t.Fatal("lease was not released after execution exited")
		}
	})
}

func TestLaunchLeaseFailure(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	for _, tt := range []struct {
		name     string
		acquired bool
		err      error
	}{
		{name: "storage error", err: wantErr},
		{name: "lease held"},
		{name: "missing credential", acquired: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &launchStore{
				acquire: func(_ context.Context, _, _ string, ttl time.Duration) (*lease, bool, error) {
					if ttl != defaultLeaseTTL {
						t.Error("default lease TTL was not applied")
					}
					return nil, tt.acquired, tt.err
				},
				release: func(context.Context, string, string, int64) error {
					t.Error("execution must not start without a lease credential")
					return nil
				},
			}
			s := &service{executions: &executionRegistry{}, store: store}
			err := s.launch(context.Background(), Task{}, run{ID: "run_1"})
			if err == nil || (tt.err != nil && !errors.Is(err, tt.err)) {
				t.Fatalf("unexpected launch error: %v", err)
			}
		})
	}
}

func TestLaunchCancelledBeforeAcquire(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := &service{executions: &executionRegistry{}, store: &launchStore{acquire: func(context.Context, string, string, time.Duration) (*lease, bool, error) {
		t.Fatal("lease acquired after request cancellation")
		return nil, false, nil
	}}}
	if err := s.launch(ctx, Task{}, run{ID: "run_1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}

func TestLaunchStatusTransitions(t *testing.T) {
	for _, tt := range []struct {
		name        string
		startStatus RunStatus
		runErr      error
		panicRun    bool
		rejectStart bool
		startErr    error
		leaseLost   bool
		want        []RunStatus
	}{
		{name: "normal return handled by recorder", startStatus: RunStatusPending, want: []RunStatus{RunStatusRunning}},
		{name: "failure", startStatus: RunStatusPending, runErr: errors.New("model unavailable"), want: []RunStatus{RunStatusRunning, RunStatusFailed}},
		{name: "cancelled", startStatus: RunStatusPending, runErr: context.Canceled, want: []RunStatus{RunStatusRunning, RunStatusCancelled}},
		{name: "panic", startStatus: RunStatusPending, panicRun: true, want: []RunStatus{RunStatusRunning, RunStatusFailed}},
		{name: "resume", startStatus: RunStatusWaitingForApproval, want: []RunStatus{RunStatusRunning}},
		{name: "start condition mismatch", startStatus: RunStatusPending, rejectStart: true, want: []RunStatus{RunStatusRunning}},
		{name: "start storage error", startStatus: RunStatusPending, startErr: errors.New("storage unavailable"), want: []RunStatus{RunStatusRunning}},
		{name: "lease lost", startStatus: RunStatusPending, leaseLost: true, want: []RunStatus{RunStatusRunning}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var statuses []RunStatus
				called, released := false, false
				store := &launchStore{
					acquire: func(context.Context, string, string, time.Duration) (*lease, bool, error) {
						return &lease{Owner: "executor_1", Epoch: 9}, true, nil
					},
					release: func(context.Context, string, string, int64) error {
						released = true
						return nil
					},
					transition: func(ctx context.Context, id string, request UpdateRunStatusRequest) (bool, error) {
						statuses = append(statuses, request.To)
						if released || ctx.Err() != nil || id != "run_1" || request.Execution == nil || request.Execution.Owner != "executor_1" || request.Execution.Epoch != 9 {
							t.Error("status update must happen before release with a usable context and acquired credential")
						}
						if request.To == RunStatusRunning {
							if !reflect.DeepEqual(request.From, []RunStatus{tt.startStatus}) {
								t.Error("unexpected source status")
							}
							return !tt.rejectStart, tt.startErr
						}
						if _, ok := ctx.Deadline(); !ok {
							t.Error("failure cleanup must have a timeout")
						}
						return true, nil
					},
				}
				s := &service{executions: &executionRegistry{}, store: store, runner: launchRunnerFunc(func(ctx context.Context, _ agentloop.Input, _ llm.Model, _ agentloop.ToolExecutor, _ agentloop.Recorder) error {
					called = true
					if tt.panicRun {
						panic("broken runner")
					}
					if tt.leaseLost {
						<-ctx.Done()
						return ctx.Err()
					}
					return tt.runErr
				})}
				if err := s.launch(context.Background(), Task{ID: "task_1"}, run{ID: "run_1", Status: tt.startStatus}); err != nil {
					t.Fatal(err)
				}
				synctest.Wait()
				if tt.leaseLost {
					time.Sleep(defaultLeaseTTL / 3)
					synctest.Wait()
				}
				if !reflect.DeepEqual(statuses, tt.want) || !released {
					t.Fatalf("statuses=%v, want=%v, released=%v", statuses, tt.want, released)
				}
				if called != (!tt.rejectStart && tt.startErr == nil) {
					t.Error("runner start did not respect state update result")
				}
			})
		})
	}
}
