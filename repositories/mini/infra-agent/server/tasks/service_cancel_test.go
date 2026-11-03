package tasks

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
)

type cancellationStore struct {
	stubTaskStore
	item      *run
	err       error
	updateErr error
	updates   []UpdateRunStatusRequest
}

func (s *cancellationStore) GetRun(ctx context.Context, _ string) (*run, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return s.item, s.err
}

func (s *cancellationStore) UpdateRunStatus(ctx context.Context, _ string, request UpdateRunStatusRequest) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	s.updates = append(s.updates, request)
	if s.updateErr != nil {
		return false, s.updateErr
	}
	for _, from := range request.From {
		if s.item != nil && s.item.Status == from {
			s.item.Status = request.To
			return true, nil
		}
	}
	return false, nil
}

func TestCancelInactiveRun(t *testing.T) {
	for _, status := range []RunStatus{RunStatusPending, RunStatusRunning, RunStatusWaitingForApproval, RunStatusWaitingForInput, RunStatusCompleted, RunStatusFailed, RunStatusCancelled} {
		t.Run(string(status), func(t *testing.T) {
			storage := &cancellationStore{item: &run{ID: "run_1", Status: status}}
			s := &service{store: storage, executions: &executionRegistry{}}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := s.CancelRun(ctx, "run_1"); err != nil {
				t.Fatal(err)
			}
			if status.IsTerminal() {
				if storage.item.Status != status || len(storage.updates) != 0 {
					t.Fatal("terminal run was modified")
				}
			} else if storage.item.Status != RunStatusCancelled || len(storage.updates) != 1 || storage.updates[0].Execution != nil {
				t.Fatalf("inactive cancellation must not carry an execution credential: %+v", storage.updates)
			}
			if err := s.CancelRun(context.Background(), "run_1"); err != nil {
				t.Fatalf("repeat cancellation: %v", err)
			}
		})
	}
}

func TestCancelRunErrors(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	for _, storage := range []*cancellationStore{
		{},
		{err: wantErr},
		{item: &run{ID: "run_1", Status: RunStatusWaitingForInput}, updateErr: wantErr},
		{item: &run{ID: "run_1", Status: RunStatus("unknown")}},
	} {
		s := &service{store: storage, executions: &executionRegistry{}}
		if err := s.CancelRun(context.Background(), "run_1"); err == nil {
			t.Fatal("expected cancellation error")
		} else if (storage.err != nil || storage.updateErr != nil) && !errors.Is(err, wantErr) {
			t.Fatalf("lost storage error: %v", err)
		}
	}
}

func TestCancelActiveRunUsesLaunchCleanup(t *testing.T) {
	for _, beforeStart := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			storage := &cancellationStore{item: &run{ID: "run_1", Status: RunStatusPending}}
			released := false
			store := &launchStore{
				acquire: func(context.Context, string, string, time.Duration) (*lease, bool, error) {
					return &lease{Owner: "executor_1", Epoch: 1}, true, nil
				},
				release:    func(context.Context, string, string, int64) error { released = true; return nil },
				transition: storage.UpdateRunStatus,
			}
			s := &service{store: store, executions: &executionRegistry{}, runner: launchRunnerFunc(func(ctx context.Context, _ agentloop.Input, _ llm.Model, _ agentloop.ToolExecutor, _ agentloop.Recorder) error {
				<-ctx.Done()
				// 即使循环恰好正常返回，用户取消也不能被遗漏。
				return nil
			})}
			if err := s.launch(context.Background(), Task{}, *storage.item); err != nil {
				t.Fatal(err)
			}
			if !beforeStart {
				synctest.Wait()
			}
			if err := s.CancelRun(context.Background(), "run_1"); err != nil {
				t.Fatal(err)
			}
			synctest.Wait()
			if !released || storage.item.Status != RunStatusCancelled || s.executions.Cancel("run_1") {
				t.Fatalf("released=%v, status=%s", released, storage.item.Status)
			}
			last := storage.updates[len(storage.updates)-1]
			if last.Execution == nil || last.Execution.Owner != "executor_1" || last.Execution.Epoch != 1 {
				t.Fatal("launch cleanup lost lease credential")
			}
		})
	}
}

func TestLaunchRejectsCancelledRun(t *testing.T) {
	s := &service{store: &cancellationStore{item: &run{ID: "run_1", Status: RunStatusCancelled}}, executions: &executionRegistry{}}
	if err := s.launch(context.Background(), Task{}, run{ID: "run_1", Status: RunStatusPending}); err == nil {
		t.Fatal("cancelled run was restarted")
	}
}

func TestExecutionRegistryIdentity(t *testing.T) {
	r := &executionRegistry{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	item, err := r.Register("run_1", cancel)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Register("run_1", cancel); err == nil {
		t.Fatal("duplicate execution accepted")
	}
	r.Unregister("run_1", &execution{})
	if !r.Cancel("run_1") || ctx.Err() == nil {
		t.Fatal("wrong execution was removed or cancellation not delivered")
	}
	r.Unregister("run_1", item)
	if r.Cancel("run_1") {
		t.Fatal("execution was not removed")
	}
}
