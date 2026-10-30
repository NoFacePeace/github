package tasks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

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

type stubTaskStore struct {
	createTaskCalls               int
	task                          Task
	err                           error
	runs                          []run
	listRunsErr                   error
	listRunsCalls                 int
	listRunsCtx                   context.Context
	listRunsRequest               ListRunsRequest
	createRunWithAttachmentsCalls int
	createRunWithAttachmentsCtx   context.Context
	createdRun                    run
	createRunWithAttachmentsErr   error
	runTaskID                     string
	attachmentIDs                 []string
}

func (s *stubTaskStore) CreateTask(_ context.Context, task Task) error {
	s.createTaskCalls++
	s.task = task
	return s.err
}

func (s *stubTaskStore) CreateRun(_ context.Context, _ run) error {
	return nil
}

func (s *stubTaskStore) CreateRunWithAttachments(ctx context.Context, taskID string, item run, attachmentIDs []string) error {
	s.createRunWithAttachmentsCalls++
	s.createRunWithAttachmentsCtx = ctx
	s.createdRun = item
	s.runTaskID = taskID
	s.attachmentIDs = attachmentIDs
	return s.createRunWithAttachmentsErr
}

func (s *stubTaskStore) ListRuns(ctx context.Context, request ListRunsRequest) ([]run, error) {
	s.listRunsCalls++
	s.listRunsCtx = ctx
	s.listRunsRequest = request
	return s.runs, s.listRunsErr
}

func (s *stubTaskStore) UpdateTask(_ context.Context, _ string, _ UpdateTaskRequest) error {
	return nil
}

func (s *stubTaskStore) AppendEvent(_ context.Context, _ Event) error {
	return nil
}

func TestNewTaskServiceInjectsDependencies(t *testing.T) {
	workspaceService := &stubWorkspaceService{}
	workspaceConfig := workspaces.Config{DataDirectory: t.TempDir()}
	cfg := Config{DefaultModel: "provider/model", TitleTimeout: 5 * time.Second}
	modelManager := &stubModelManager{}
	providerManager := &stubProviderManager{}
	taskService := NewTaskService(workspaceService, workspaceConfig, cfg, modelManager, providerManager).(*service)
	if taskService.workspaceService != workspaceService {
		t.Fatal("NewTaskService() did not retain the workspace service")
	}
	if taskService.workspaceConfig != workspaceConfig {
		t.Fatal("NewTaskService() did not retain the workspace config")
	}
	if taskService.config != cfg {
		t.Fatal("NewTaskService() did not retain the task config")
	}
	if taskService.modelManager != modelManager {
		t.Fatal("NewTaskService() did not retain the model manager")
	}
	if taskService.providerManager != providerManager {
		t.Fatal("NewTaskService() did not retain the provider manager")
	}
}

func TestNewTaskServiceTitleTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{name: "unconfigured", want: 30 * time.Second},
		{name: "negative", timeout: -time.Second, want: 30 * time.Second},
		{name: "configured", timeout: 5 * time.Second, want: 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{TitleTimeout: tt.timeout}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, cfg, &stubModelManager{}, &stubProviderManager{}).(*service)
			if got := taskService.config.TitleTimeout; got != tt.want {
				t.Errorf("TitleTimeout = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceCreateSetsGoal(t *testing.T) {
	const goal = "Add model enable support"
	attachmentIDs := []string{"attachment_1", "attachment_2"}
	taskStore := &stubTaskStore{}
	taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
	taskService.store = taskStore

	task, err := taskService.Create(context.Background(), CreateTaskRequest{
		Goal:               goal,
		AttachmentIDs:      attachmentIDs,
		WorkspaceDirectory: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.Goal != goal {
		t.Errorf("Create() goal = %q, want %q", task.Goal, goal)
	}
	if taskStore.createTaskCalls != 1 {
		t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
	}
	if taskStore.task.Goal != goal {
		t.Errorf("stored task goal = %q, want %q", taskStore.task.Goal, goal)
	}
	assertPrefixedUUIDv7(t, task.ID, "task")
	if taskStore.task.ID != task.ID {
		t.Errorf("stored task ID = %q, want %q", taskStore.task.ID, task.ID)
	}
	if taskStore.listRunsRequest.TaskID != task.ID {
		t.Errorf("ListRuns() task ID = %q, want %q", taskStore.listRunsRequest.TaskID, task.ID)
	}
	assertPrefixedUUIDv7(t, taskStore.createdRun.ID, "run")
	if taskStore.runTaskID != task.ID {
		t.Errorf("CreateRunWithAttachments() task ID = %q, want %q", taskStore.runTaskID, task.ID)
	}
	if !slices.Equal(taskStore.attachmentIDs, attachmentIDs) {
		t.Errorf("CreateRunWithAttachments() attachment IDs = %v, want %v", taskStore.attachmentIDs, attachmentIDs)
	}
}

func TestServiceCreateSetsTitleFromGoal(t *testing.T) {
	tests := []struct {
		name          string
		goal          string
		maxTitleRunes int
		want          string
	}{
		{name: "goal below limit", goal: "Goal text", maxTitleRunes: 10, want: "Goal text"},
		{name: "truncate goal", goal: "Goal text", maxTitleRunes: 5, want: "Go..."},
		{name: "empty goal", maxTitleRunes: 5},
		{name: "unconfigured limit", goal: "Goal text", want: "Goal text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStore := &stubTaskStore{}
			cfg := Config{MaxTitleRunes: tt.maxTitleRunes}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, cfg, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

			task, err := taskService.Create(context.Background(), CreateTaskRequest{
				Goal:               tt.goal,
				WorkspaceDirectory: t.TempDir(),
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if task.Title != tt.want {
				t.Errorf("Create() title = %q, want %q", task.Title, tt.want)
			}
			if taskStore.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
			}
			if taskStore.task.Title != tt.want {
				t.Errorf("stored task title = %q, want %q", taskStore.task.Title, tt.want)
			}
			if task.Goal != tt.goal || taskStore.task.Goal != tt.goal {
				t.Errorf("Create() goal = %q, stored goal = %q, want %q", task.Goal, taskStore.task.Goal, tt.goal)
			}
		})
	}
}

func TestServiceCreateAutoTitleKeepsFallback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		taskStore := &stubTaskStore{}
		cfg := Config{AutoTitle: true, MaxTitleRunes: 5}
		taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, cfg, &stubModelManager{}, &stubProviderManager{}).(*service)
		taskService.store = taskStore

		task, err := taskService.Create(ctx, CreateTaskRequest{
			Goal:               "Goal text",
			WorkspaceDirectory: t.TempDir(),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		cancel()
		synctest.Wait()

		const want = "Go..."
		if task.Title != want || taskStore.task.Title != want {
			t.Errorf("task title = %q, stored title = %q, want %q", task.Title, taskStore.task.Title, want)
		}
		if taskStore.createTaskCalls != 1 {
			t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
		}
	})
}

func TestServiceCreateGetsWorkspace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workspaceService := &stubWorkspaceService{status: workspaces.StatusActive, directory: t.TempDir()}
	taskStore := &stubTaskStore{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
	taskService.store = taskStore
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
	if taskStore.createTaskCalls != 1 {
		t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
	}
	want := Task{
		ID:                 task.ID,
		TriggerType:        TriggerTypeManual,
		AgentID:            agents.DefaultAgentID,
		WorkspaceID:        "workspace_123",
		WorkspaceDirectory: workspaceService.directory,
	}
	if task != want || taskStore.task != want {
		t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStore.task, want)
	}
}

func TestServiceCreateWrapsStoreError(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	taskStore := &stubTaskStore{err: wantErr}
	taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
	taskService.store = taskStore

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
	if taskStore.createTaskCalls != 1 {
		t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
	}
}

func TestServiceCreateReturnsWorkspaceError(t *testing.T) {
	wantErr := errors.New("workspace not found")
	workspaceService := &stubWorkspaceService{err: wantErr}
	taskStore := &stubTaskStore{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
	taskService.store = taskStore

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
	if taskStore.createTaskCalls != 0 {
		t.Errorf("CreateTask() calls = %d, want 0", taskStore.createTaskCalls)
	}
}

func TestServiceCreateRejectsArchivedWorkspace(t *testing.T) {
	workspaceService := &stubWorkspaceService{status: workspaces.StatusArchived}
	taskStore := &stubTaskStore{}
	taskService := NewTaskService(workspaceService, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
	taskService.store = taskStore

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
	if taskStore.createTaskCalls != 0 {
		t.Errorf("CreateTask() calls = %d, want 0", taskStore.createTaskCalls)
	}
}

func TestServiceCreateUsesWorkspaceDirectory(t *testing.T) {
	directory := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(directory, link); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		directory string
	}{
		{name: "directory", directory: directory},
		{name: "symlink", directory: link},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStore := &stubTaskStore{}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

			task, err := taskService.Create(context.Background(), CreateTaskRequest{
				WorkspaceDirectory: tt.directory,
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			want := Task{
				ID:                 task.ID,
				TriggerType:        TriggerTypeManual,
				AgentID:            agents.DefaultAgentID,
				WorkspaceDirectory: tt.directory,
			}
			if task != want || taskStore.task != want {
				t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStore.task, want)
			}
			if taskStore.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
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
			taskStore := &stubTaskStore{}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

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
			if taskStore.createTaskCalls != 0 {
				t.Errorf("CreateTask() calls = %d, want 0", taskStore.createTaskCalls)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := workspaces.Config{DataDirectory: filepath.Join(t.TempDir(), "data")}
			taskStore := &stubTaskStore{}
			taskService := NewTaskService(&stubWorkspaceService{}, cfg, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

			task, err := taskService.Create(context.Background(), CreateTaskRequest{WorkspaceID: tt.workspaceID})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			want := Task{
				ID:                 task.ID,
				TriggerType:        TriggerTypeManual,
				AgentID:            agents.DefaultAgentID,
				WorkspaceID:        tt.wantID,
				WorkspaceDirectory: filepath.Join(cfg.DataDirectory, "workspaces", tt.wantID),
			}
			if task != want || taskStore.task != want {
				t.Errorf("Create() task = %+v, stored task = %+v, want %+v", task, taskStore.task, want)
			}
			if taskStore.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
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
		{name: "file", cfg: workspaces.Config{DataDirectory: file}, wantFilesystemErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskStore := &stubTaskStore{}
			taskService := NewTaskService(&stubWorkspaceService{}, tt.cfg, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

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
			if taskStore.createTaskCalls != 0 {
				t.Errorf("CreateTask() calls = %d, want 0", taskStore.createTaskCalls)
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
			taskStore := &stubTaskStore{}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = taskStore

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
			if taskStore.task.AgentID != tt.want {
				t.Errorf("stored Task.AgentID = %q, want %q", taskStore.task.AgentID, tt.want)
			}
			if taskStore.createTaskCalls != 1 {
				t.Errorf("CreateTask() calls = %d, want 1", taskStore.createTaskCalls)
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
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, &stubModelManager{}, &stubProviderManager{}).(*service)
			taskService.store = &stubTaskStore{}
			task, err := taskService.Create(context.Background(), CreateTaskRequest{
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
