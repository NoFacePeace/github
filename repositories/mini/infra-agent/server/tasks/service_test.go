package tasks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"infra-agent/server/agents"
	"infra-agent/server/workspaces"
)

type stubWorkspaceService struct {
	ctx       context.Context
	id        string
	calls     int
	err       error
	status    workspaces.Status
	directory string
}

func (s *stubWorkspaceService) Get(ctx context.Context, id string) (workspaces.Workspace, error) {
	s.ctx = ctx
	s.id = id
	s.calls++
	return workspaces.Workspace{ID: id, Directory: s.directory, Status: s.status}, s.err
}

type stubTaskStorage struct {
	createTaskCalls int
	task            Task
	err             error
}

func (s *stubTaskStorage) CreateTask(_ context.Context, task Task) error {
	s.createTaskCalls++
	s.task = task
	return s.err
}

func (s *stubTaskStorage) CreateRun(_ context.Context, _ run) error {
	return nil
}

func TestNewTaskServiceInjectsDependencies(t *testing.T) {
	workspaceService := &stubWorkspaceService{}
	workspaceConfig := workspaces.Config{DataDirectory: t.TempDir()}
	cfg := Config{DefaultModel: "provider/model"}
	taskService := NewTaskService(workspaceService, workspaceConfig, cfg).(*service)
	if taskService.workspaceService != workspaceService {
		t.Fatal("NewTaskService() did not retain the workspace service")
	}
	if taskService.workspaceConfig != workspaceConfig {
		t.Fatal("NewTaskService() did not retain the workspace config")
	}
	if taskService.config != cfg {
		t.Fatal("NewTaskService() did not retain the task config")
	}
}

func TestServiceCreateGetsWorkspace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workspaceService := &stubWorkspaceService{status: workspaces.StatusActive, directory: t.TempDir()}
	taskStorage := &stubTaskStorage{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}).(*service)
	taskService.manager.storage = taskStorage
	task, err := taskService.Create(ctx, CreateTaskRequest{
		WorkspaceID:        "workspace_123",
		WorkspaceDirectory: "ignored-relative-directory",
	})
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
	want := Task{
		TriggerType:        TriggerTypeManual,
		AgentID:            agents.DefaultAgentID,
		WorkspaceID:        "workspace_123",
		WorkspaceDirectory: workspaceService.directory,
	}
	if task != want || taskStorage.task != want {
		t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStorage.task, want)
	}
}

func TestServiceCreateWrapsManagerError(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	taskStorage := &stubTaskStorage{err: wantErr}
	taskService := NewTaskService(nil, workspaces.Config{}, Config{}).(*service)
	taskService.manager.storage = taskStorage

	task, err := taskService.Create(context.Background(), CreateTaskRequest{
		WorkspaceDirectory: t.TempDir(),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want wrapped %v", err, wantErr)
	}
	if want := "create task: storage unavailable"; err.Error() != want {
		t.Errorf("Create() error message = %q, want %q", err.Error(), want)
	}
	if task != (Task{}) {
		t.Errorf("Create() task = %+v, want empty task", task)
	}
	if taskStorage.createTaskCalls != 1 {
		t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
	}
}

func TestServiceCreateReturnsWorkspaceError(t *testing.T) {
	wantErr := errors.New("workspace not found")
	workspaceService := &stubWorkspaceService{err: wantErr}
	taskStorage := &stubTaskStorage{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}).(*service)
	taskService.manager.storage = taskStorage

	_, err := taskService.Create(context.Background(), CreateTaskRequest{
		WorkspaceID:        "missing",
		WorkspaceDirectory: t.TempDir(),
	})
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

func TestServiceCreateRejectsArchivedWorkspace(t *testing.T) {
	workspaceService := &stubWorkspaceService{status: workspaces.StatusArchived}
	taskStorage := &stubTaskStorage{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}).(*service)
	taskService.manager.storage = taskStorage

	_, err := taskService.Create(context.Background(), CreateTaskRequest{
		WorkspaceID:        "workspace_123",
		WorkspaceDirectory: t.TempDir(),
	})
	if err == nil {
		t.Fatal("Create() error = nil, want archived workspace error")
	}
	if want := `workspace "workspace_123" is archived`; err.Error() != want {
		t.Errorf("Create() error message = %q, want %q", err.Error(), want)
	}
	if taskStorage.createTaskCalls != 0 {
		t.Errorf("CreateTask() calls = %d, want 0", taskStorage.createTaskCalls)
	}
}

func TestServiceCreateUsesWorkspaceDirectory(t *testing.T) {
	directory := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(directory, link); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		directory   string
		workspaceID string
	}{
		{name: "directory", directory: directory, workspaceID: "workspace_123"},
		{name: "without id", directory: directory},
		{name: "symlink", directory: link, workspaceID: "workspace_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, workspaces.Config{}, Config{}).(*service)
			taskService.manager.storage = taskStorage

			task, err := taskService.Create(context.Background(), CreateTaskRequest{
				WorkspaceID:        tt.workspaceID,
				WorkspaceDirectory: tt.directory,
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			want := Task{
				TriggerType:        TriggerTypeManual,
				AgentID:            agents.DefaultAgentID,
				WorkspaceID:        tt.workspaceID,
				WorkspaceDirectory: tt.directory,
			}
			if task != want || taskStorage.task != want {
				t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStorage.task, want)
			}
			if taskStorage.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
			}
		})
	}
}

func TestServiceCreateRejectsWorkspaceDirectory(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(directory, "missing")
	tests := []struct {
		name         string
		directory    string
		wantNotExist bool
	}{
		{name: "relative", directory: "workspace"},
		{name: "missing", directory: missing, wantNotExist: true},
		{name: "file", directory: file},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, workspaces.Config{}, Config{}).(*service)
			taskService.manager.storage = taskStorage

			_, err := taskService.Create(context.Background(), CreateTaskRequest{WorkspaceDirectory: tt.directory})
			if err == nil {
				t.Fatal("Create() error = nil, want directory validation error")
			}
			if !strings.HasPrefix(err.Error(), "validate workspace directory: ") {
				t.Errorf("Create() error = %v, want directory validation context", err)
			}
			if tt.wantNotExist && !errors.Is(err, os.ErrNotExist) {
				t.Errorf("Create() error = %v, want wrapped os.ErrNotExist", err)
			}
			if taskStorage.createTaskCalls != 0 {
				t.Errorf("CreateTask() calls = %d, want 0", taskStorage.createTaskCalls)
			}
		})
	}
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Stat(%q) error = %v, want os.ErrNotExist", missing, err)
	}
}

func TestServiceCreateEnsuresWorkspaceDirectory(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		wantID      string
	}{
		{name: "default", wantID: "default"},
		{name: "custom", workspaceID: "workspace_123", wantID: "workspace_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := workspaces.Config{DataDirectory: filepath.Join(t.TempDir(), "data")}
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, cfg, Config{}).(*service)
			taskService.manager.storage = taskStorage

			task, err := taskService.Create(context.Background(), CreateTaskRequest{WorkspaceID: tt.workspaceID})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			want := Task{
				TriggerType:        TriggerTypeManual,
				AgentID:            agents.DefaultAgentID,
				WorkspaceID:        tt.wantID,
				WorkspaceDirectory: filepath.Join(cfg.DataDirectory, "workspaces", tt.wantID),
			}
			if task != want || taskStorage.task != want {
				t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStorage.task, want)
			}
			if taskStorage.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
			}
			info, err := os.Stat(want.WorkspaceDirectory)
			if err != nil {
				t.Fatalf("Stat(%q) error = %v", want.WorkspaceDirectory, err)
			}
			if !info.IsDir() {
				t.Errorf("Create() did not create a directory at %q", want.WorkspaceDirectory)
			}
		})
	}
}

func TestServiceCreateReturnsEnsureDirectoryError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name              string
		cfg               workspaces.Config
		workspaceID       string
		wantFilesystemErr bool
	}{
		{name: "empty config"},
		{name: "relative data directory", cfg: workspaces.Config{DataDirectory: "data"}},
		{name: "invalid id", cfg: workspaces.Config{DataDirectory: t.TempDir()}, workspaceID: "../outside"},
		{name: "file", cfg: workspaces.Config{DataDirectory: file}, wantFilesystemErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, tt.cfg, Config{}).(*service)
			taskService.manager.storage = taskStorage

			_, err := taskService.Create(context.Background(), CreateTaskRequest{WorkspaceID: tt.workspaceID})
			if err == nil {
				t.Fatal("Create() error = nil, want directory creation error")
			}
			if !strings.HasPrefix(err.Error(), "ensure workspace directory: ") {
				t.Errorf("Create() error = %v, want directory creation context", err)
			}
			var pathErr *os.PathError
			if tt.wantFilesystemErr && !errors.As(err, &pathErr) {
				t.Errorf("Create() error = %v, want wrapped *os.PathError", err)
			}
			if taskStorage.createTaskCalls != 0 {
				t.Errorf("CreateTask() calls = %d, want 0", taskStorage.createTaskCalls)
			}
		})
	}
}

func TestServiceCreateAgentID(t *testing.T) {
	tests := []struct {
		name    string
		agentID string
		want    string
	}{
		{name: "unset", want: "default"},
		{name: "specified", agentID: "agent_123", want: "agent_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, workspaces.Config{}, Config{}).(*service)
			taskService.manager.storage = taskStorage

			task, err := taskService.Create(context.Background(), CreateTaskRequest{
				AgentID:            tt.agentID,
				WorkspaceDirectory: t.TempDir(),
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if task.AgentID != tt.want {
				t.Errorf("Task.AgentID = %q, want %q", task.AgentID, tt.want)
			}
			if taskStorage.task.AgentID != tt.want {
				t.Errorf("stored Task.AgentID = %q, want %q", taskStorage.task.AgentID, tt.want)
			}
			if taskStorage.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
			}
		})
	}
}

func TestServiceCreateModel(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		defaultModel string
		want         string
	}{
		{name: "default", defaultModel: "provider/default-model", want: "provider/default-model"},
		{name: "specified", model: "provider/custom-model", defaultModel: "provider/default-model", want: "provider/custom-model"},
		{name: "specified without default", model: "provider/custom-model", want: "provider/custom-model"},
		{name: "both empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStorage := &stubTaskStorage{}
			taskService := NewTaskService(nil, workspaces.Config{}, Config{DefaultModel: tt.defaultModel}).(*service)
			taskService.manager.storage = taskStorage

			task, err := taskService.Create(context.Background(), CreateTaskRequest{
				Model:              tt.model,
				WorkspaceDirectory: t.TempDir(),
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if task.Model != tt.want {
				t.Errorf("Task.Model = %q, want %q", task.Model, tt.want)
			}
			if taskStorage.task.Model != tt.want {
				t.Errorf("stored Task.Model = %q, want %q", taskStorage.task.Model, tt.want)
			}
			if taskStorage.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStorage.createTaskCalls)
			}
		})
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
			task, err := NewTaskService(nil, workspaces.Config{}, Config{}).Create(context.Background(), CreateTaskRequest{
				TriggerType:        tt.triggerType,
				WorkspaceDirectory: t.TempDir(),
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
