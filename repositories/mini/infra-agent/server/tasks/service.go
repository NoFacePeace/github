package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"infra-agent/server/agents"
	"infra-agent/server/providers"
	"infra-agent/server/workspaces"
)

// Service 定义任务模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateTaskRequest) (Task, error)
}

// CreateTaskRequest 表示创建任务的请求。
type CreateTaskRequest struct {
	Goal               string
	TriggerType        TriggerType
	AgentID            string
	Model              string
	AttachmentIDs      []string
	WorkspaceID        string
	WorkspaceDirectory string
}

type service struct {
	executorID       string
	executions       *executionRegistry
	config           Config
	store            store
	workspaceService workspaces.Service
	workspaceConfig  workspaces.Config
	modelManager     providers.ModelManager
	providerManager  providers.Manager
}

// NewTaskService 创建任务服务，并为该执行器实例生成唯一标识。
func NewTaskService(
	workspaceService workspaces.Service,
	workspaceConfig workspaces.Config,
	cfg Config,
	modelManager providers.ModelManager,
	providerManager providers.Manager,
) (Service, error) {
	executorID, err := newID("executor")
	if err != nil {
		return nil, fmt.Errorf("generate executor id: %w", err)
	}
	if cfg.TitleTimeout <= 0 {
		cfg.TitleTimeout = defaultTitleTimeout
	}
	return &service{
		executorID: executorID,
		executions: &executionRegistry{
			active: make(map[string]*execution),
		},
		config:           cfg,
		workspaceService: workspaceService,
		workspaceConfig:  workspaceConfig,
		modelManager:     modelManager,
		providerManager:  providerManager,
	}, nil
}

// Create 创建一个任务。
func (s *service) Create(ctx context.Context, request CreateTaskRequest) (Task, error) {
	if request.TriggerType == "" {
		request.TriggerType = TriggerTypeManual
	}
	workspaceID := request.WorkspaceID
	workspaceDirectory := request.WorkspaceDirectory
	if workspaceID != "" {
		workspace, err := s.workspaceService.Get(ctx, request.WorkspaceID)
		if err != nil {
			return Task{}, fmt.Errorf("get workspace %q: %w", request.WorkspaceID, err)
		}
		if workspace.Status == workspaces.StatusArchived {
			return Task{}, fmt.Errorf("workspace %q is archived", request.WorkspaceID)
		}
		workspaceID = workspace.ID
		workspaceDirectory = workspace.Directory
	} else if workspaceDirectory != "" {
		if err := workspaces.ValidateDirectory(workspaceDirectory); err != nil {
			return Task{}, fmt.Errorf("validate workspace directory: %w", err)
		}
	} else {
		workspaceID = workspaces.DefaultWorkspaceID
		var err error
		workspaceDirectory, err = workspaces.EnsureDirectory(s.workspaceConfig, workspaceID)
		if err != nil {
			return Task{}, fmt.Errorf("ensure workspace directory: %w", err)
		}
	}
	if request.AgentID == "" {
		request.AgentID = agents.DefaultAgentID
	}
	taskID, err := newID("task")
	if err != nil {
		return Task{}, fmt.Errorf("generate task id: %w", err)
	}
	task := Task{
		ID:                 taskID,
		Goal:               request.Goal,
		Title:              truncateRunes(request.Goal, s.config.MaxTitleRunes),
		TriggerType:        request.TriggerType,
		AgentID:            request.AgentID,
		WorkspaceID:        workspaceID,
		WorkspaceDirectory: workspaceDirectory,
	}
	if err := s.store.CreateTask(ctx, task); err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	if request.Model == "" {
		request.Model = s.config.DefaultModel
	}
	if _, err := s.startRun(ctx, task, request.AttachmentIDs); err != nil {
		return Task{}, fmt.Errorf("start run: %w", err)
	}
	if s.config.AutoTitle {
		go func(taskID, goal, model string) {
			// 保留请求上下文中的值，但使用独立的取消生命周期和超时。
			titleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.config.TitleTimeout)
			defer cancel()

			title, err := s.generateTitle(titleCtx, goal, model)
			if err != nil {
				slog.ErrorContext(titleCtx, "generate task title", "task_id", taskID, "error", err)
				return
			}
			title = strings.TrimSpace(title)
			if title == "" {
				return
			}
			title = truncateRunes(title, s.config.MaxTitleRunes)
			if err := s.store.UpdateTask(titleCtx, taskID, UpdateTaskRequest{Title: &title}); err != nil {
				slog.ErrorContext(titleCtx, "update task title", "task_id", taskID, "error", err)
			}
		}(task.ID, task.Goal, request.Model)
	}
	return task, nil
}

// startRun 为任务创建并启动一次运行。
func (s *service) startRun(ctx context.Context, task Task, attachmentIDs []string) (run, error) {
	runs, err := s.store.ListRuns(ctx, ListRunsRequest{TaskID: task.ID})
	if err != nil {
		return run{}, fmt.Errorf("list runs for task %q: %w", task.ID, err)
	}
	if len(runs) > 0 {
		latest := runs[len(runs)-1]
		if !latest.Status.IsTerminal() {
			return run{}, fmt.Errorf("latest run for task %q is not terminal: %q", task.ID, latest.Status)
		}
	}
	runID, err := newID("run")
	if err != nil {
		return run{}, fmt.Errorf("generate run id: %w", err)
	}
	item := run{ID: runID, Status: RunStatusPending}
	if err := s.store.CreateRunWithAttachments(ctx, task.ID, item, attachmentIDs); err != nil {
		return run{}, fmt.Errorf("create run with attachments: %w", err)
	}
	// TODO: 提交成功后启动执行。
	return item, nil
}

// launch 启动已创建的运行。
func (s *service) launch(ctx context.Context, task Task, item run) error {
	// TODO: 登记执行状态并启动运行。
	return nil
}

// generateTitle 根据任务目标生成标题，不更新任务。
func (s *service) generateTitle(ctx context.Context, goal, model string) (string, error) {
	// TODO: 调用模型生成任务标题。
	return "", nil
}

// resolveModelMaxOutputTokens 根据模型标识确定最大输出 token 数。
// 返回 0 且无错误仅表示匹配到的启用模型未配置上限。
func (s *service) resolveModelMaxOutputTokens(ctx context.Context, model string) (int, error) {
	providerName, modelID, ok := strings.Cut(model, "/")
	if !ok || providerName == "" || modelID == "" {
		return 0, fmt.Errorf("invalid model identifier %q: expected provider/model", model)
	}
	models, err := s.modelManager.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list models: %w", err)
	}
	for _, m := range models {
		if !m.Enabled || m.ModelID != modelID {
			continue
		}
		provider, err := s.providerManager.Get(ctx, m.ProviderID)
		if err != nil {
			return 0, fmt.Errorf("get provider %q: %w", m.ProviderID, err)
		}
		if provider.Name == providerName {
			return m.MaxOutputTokens, nil
		}
	}
	return 0, fmt.Errorf("model %q not found or disabled", model)
}
