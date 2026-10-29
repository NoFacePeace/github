package tasks

import (
	"context"
	"errors"
	"testing"

	"infra-agent/server/workspaces"
)

type stubWorkspaceService struct {
	ctx   context.Context
	id    string
	calls int
	err   error
}

func (s *stubWorkspaceService) Get(ctx context.Context, id string) (workspaces.Workspace, error) {
	s.ctx = ctx
	s.id = id
	s.calls++
	return workspaces.Workspace{ID: id}, s.err
}

type stubTaskStorage struct {
	createTaskCalls int
}

func (s *stubTaskStorage) CreateTask(_ context.Context, _ Task) error {
	s.createTaskCalls++
	return nil
}

func (s *stubTaskStorage) CreateRun(_ context.Context, _ run) error {
	return nil
}

func TestNewTaskServiceInjectsWorkspaceService(t *testing.T) {
	workspaceService := &stubWorkspaceService{}
	taskService := NewTaskService(workspaceService).(*service)
	if taskService.workspaceService != workspaceService {
		t.Fatal("NewTaskService() did not retain the workspace service")
	}
}

func TestServiceCreateGetsWorkspace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workspaceService := &stubWorkspaceService{}
	taskStorage := &stubTaskStorage{}
	taskService := NewTaskService(workspaceService).(*service)
	taskService.manager.storage = taskStorage
	_, err := taskService.Create(ctx, CreateTaskRequest{WorkspaceID: "workspace_123"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if workspaceService.calls != 1 {
		t.Errorf("Get() calls = %d, want 1", workspaceService.calls)
	}
	if workspaceService.id != "workspace_123" {
		t.Errorf("Get() id = %q, want %q", workspaceService.id, "workspace_123")
	}
	if workspaceService.ctx != ctx {
		t.Error("Get() did not receive the request context")
	}
	if taskStorage.createTaskCalls != 1 {
		t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
	}
}

func TestServiceCreateReturnsWorkspaceError(t *testing.T) {
	wantErr := errors.New("workspace not found")
	workspaceService := &stubWorkspaceService{err: wantErr}
	taskStorage := &stubTaskStorage{}
	taskService := NewTaskService(workspaceService).(*service)
	taskService.manager.storage = taskStorage

	_, err := taskService.Create(context.Background(), CreateTaskRequest{WorkspaceID: "missing"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
	if want := `get workspace "missing": workspace not found`; err.Error() != want {
		t.Errorf("Create() error message = %q, want %q", err.Error(), want)
	}
	if taskStorage.createTaskCalls != 0 {
		t.Errorf("CreateTask() calls = %d, want 0", taskStorage.createTaskCalls)
	}
}

func TestServiceCreateTriggerType(t *testing.T) {
	tests := []struct {
		name        string
		triggerType TriggerType
		want        TriggerType
	}{
		{name: "unset", want: TriggerTypeManual},
		{name: "manual", triggerType: TriggerTypeManual, want: TriggerTypeManual},
		{name: "schedule", triggerType: TriggerTypeSchedule, want: TriggerTypeSchedule},
		{name: "webhook", triggerType: TriggerTypeWebhook, want: TriggerTypeWebhook},
		{name: "im", triggerType: TriggerTypeIM, want: TriggerTypeIM},
		{name: "event", triggerType: TriggerTypeEvent, want: TriggerTypeEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTaskService(nil).Create(context.Background(), CreateTaskRequest{
				TriggerType: tt.triggerType,
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if task.TriggerType != tt.want {
				t.Errorf("Task.TriggerType = %q, want %q", task.TriggerType, tt.want)
			}
		})
	}
}
