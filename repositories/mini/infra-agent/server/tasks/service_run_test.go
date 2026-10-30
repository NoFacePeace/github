package tasks

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestStartRunChecksLatestStatus(t *testing.T) {
	tests := []struct {
		name          string
		runs          []run
		wantErr       bool
		attachmentIDs []string
	}{
		{name: "no runs"},
		{name: "with attachments", attachmentIDs: []string{"attachment_1", "attachment_2"}},
		{name: "completed", runs: []run{{Status: RunStatusCompleted}}},
		{name: "failed", runs: []run{{Status: RunStatusFailed}}},
		{name: "cancelled", runs: []run{{Status: RunStatusCancelled}}},
		{name: "pending", runs: []run{{Status: RunStatusPending}}, wantErr: true},
		{name: "running", runs: []run{{Status: RunStatusRunning}}, wantErr: true},
		{name: "waiting for input", runs: []run{{Status: RunStatusWaitingForInput}}, wantErr: true},
		{name: "waiting for approval", runs: []run{{Status: RunStatusWaitingForApproval}}, wantErr: true},
		{name: "empty status", runs: []run{{}}, wantErr: true},
		{name: "unknown status", runs: []run{{Status: RunStatus("unknown")}}, wantErr: true},
		{name: "latest running", runs: []run{{Status: RunStatusCompleted}, {Status: RunStatusRunning}}, wantErr: true},
		{name: "latest completed", runs: []run{{Status: RunStatusFailed}, {Status: RunStatusCompleted}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			store := &stubTaskStore{runs: tt.runs}
			s := &service{store: store}

			item, err := s.startRun(ctx, Task{ID: "task_123"}, tt.attachmentIDs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("startRun() error = %v, want error = %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "is not terminal") {
				t.Errorf("startRun() error = %v, want non-terminal status error", err)
			}
			if tt.wantErr {
				if store.createRunWithAttachmentsCalls != 0 {
					t.Errorf("CreateRunWithAttachments() calls = %d, want 0", store.createRunWithAttachmentsCalls)
				}
			} else {
				if store.createRunWithAttachmentsCalls != 1 || store.createRunWithAttachmentsCtx != ctx {
					t.Error("CreateRunWithAttachments() must be called once with the request context")
				}
				assertPrefixedUUIDv7(t, item.ID, "run")
				want := run{ID: item.ID, Status: RunStatusPending}
				if item != want || store.createdRun != want {
					t.Errorf("startRun() run = %+v, stored run = %+v, want %+v", item, store.createdRun, want)
				}
				if store.runTaskID != "task_123" {
					t.Errorf("CreateRunWithAttachments() task ID = %q, want task_123", store.runTaskID)
				}
				if !slices.Equal(store.attachmentIDs, tt.attachmentIDs) {
					t.Errorf("CreateRunWithAttachments() attachment IDs = %v, want %v", store.attachmentIDs, tt.attachmentIDs)
				}
			}
			if store.listRunsCalls != 1 || store.listRunsCtx != ctx {
				t.Error("ListRuns() must be called once with the request context")
			}
			if store.listRunsRequest.TaskID != "task_123" {
				t.Errorf("ListRuns() task ID = %q, want task_123", store.listRunsRequest.TaskID)
			}
		})
	}
}

func TestStartRunWrapsListError(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	store := &stubTaskStore{listRunsErr: wantErr}
	s := &service{store: store}
	_, err := s.startRun(context.Background(), Task{ID: "task_123"}, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("startRun() error = %v, want wrapped %v", err, wantErr)
	}
	if want := `list runs for task "task_123": storage unavailable`; err.Error() != want {
		t.Errorf("startRun() error = %q, want %q", err.Error(), want)
	}
	if store.createRunWithAttachmentsCalls != 0 {
		t.Errorf("CreateRunWithAttachments() calls = %d, want 0", store.createRunWithAttachmentsCalls)
	}
}

func TestStartRunWrapsCreateError(t *testing.T) {
	wantErr := errors.New("store unavailable")
	store := &stubTaskStore{createRunWithAttachmentsErr: wantErr}
	s := &service{store: store}
	item, err := s.startRun(context.Background(), Task{ID: "task_123"}, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("startRun() error = %v, want wrapped %v", err, wantErr)
	}
	if want := "create run with attachments: store unavailable"; err.Error() != want {
		t.Errorf("startRun() error = %q, want %q", err.Error(), want)
	}
	if item != (run{}) {
		t.Errorf("startRun() run = %+v, want empty run", item)
	}
	if store.createRunWithAttachmentsCalls != 1 {
		t.Errorf("CreateRunWithAttachments() calls = %d, want 1", store.createRunWithAttachmentsCalls)
	}
}

func TestStartRunQueriesHistoryWithoutTaskID(t *testing.T) {
	store := &stubTaskStore{runs: []run{{Status: RunStatusRunning}}}
	s := &service{store: store}
	_, err := s.startRun(context.Background(), Task{}, nil)
	if err == nil || !strings.Contains(err.Error(), "is not terminal") {
		t.Fatalf("startRun() error = %v, want non-terminal status error", err)
	}
	if store.listRunsCalls != 1 {
		t.Errorf("ListRuns() calls = %d, want 1", store.listRunsCalls)
	}
	if store.listRunsRequest.TaskID != "" {
		t.Errorf("ListRuns() task ID = %q, want empty task ID", store.listRunsRequest.TaskID)
	}
}
