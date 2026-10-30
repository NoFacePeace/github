package tasks

import (
	"context"
	"fmt"
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
	WorkspaceID        string
	WorkspaceDirectory string
}

type service struct {
	config           Config
	manager          *manager
	workspaceService workspaces.Service
	workspaceConfig  workspaces.Config
	modelManager     providers.ModelManager
	providerManager  providers.Manager
}

// NewTaskService 创建任务服务。
func NewTaskService(
	workspaceService workspaces.Service,
	workspaceConfig workspaces.Config,
	cfg Config,
	modelManager providers.ModelManager,
	providerManager providers.Manager,
) Service {
	return &service{
		config:           cfg,
		manager:          newManager(),
		workspaceService: workspaceService,
		workspaceConfig:  workspaceConfig,
		modelManager:     modelManager,
		providerManager:  providerManager,
	}
}

// Create 创建一个任务。
func (s *service) Create(ctx context.Context, request CreateTaskRequest) (Task, error) {
	if request.TriggerType == "" {
		request.TriggerType = TriggerTypeManual
	}
	workspaceID := request.WorkspaceID
	workspaceDirectory := request.WorkspaceDirectory
	if s.workspaceService != nil {
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
		if workspaceID == "" {
			workspaceID = workspaces.DefaultWorkspaceID
		}
		var err error
		workspaceDirectory, err = workspaces.EnsureDirectory(s.workspaceConfig, workspaceID)
		if err != nil {
			return Task{}, fmt.Errorf("ensure workspace directory: %w", err)
		}
	}
	if request.AgentID == "" {
		request.AgentID = agents.DefaultAgentID
	}
	task := Task{
		Goal:               request.Goal,
		TriggerType:        request.TriggerType,
		AgentID:            request.AgentID,
		WorkspaceID:        workspaceID,
		WorkspaceDirectory: workspaceDirectory,
	}
	task, err := s.manager.CreateTask(ctx, task)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	if request.Model == "" {
		request.Model = s.config.DefaultModel
	}
	if _, err := s.startRun(ctx, task); err != nil {
		return Task{}, fmt.Errorf("start run: %w", err)
	}
	return task, nil
}

// startRun 为任务创建并启动一次运行。
func (s *service) startRun(ctx context.Context, task Task) (run, error) {
	// TODO: 创建运行、保存任务关联并启动执行。
	return run{}, nil
}

// resolveModelMaxOutputTokens 根据模型标识确定最大输出 token 数。
// 返回 0 且无错误仅表示匹配到的启用模型未配置上限。
func (s *service) resolveModelMaxOutputTokens(ctx context.Context, model string) (int, error) {
	providerName, modelID, ok := strings.Cut(model, "/")
	if !ok || providerName == "" || modelID == "" {
		return 0, fmt.Errorf("invalid model identifier %q: expected provider/model", model)
	}
	if s.modelManager == nil {
		return 0, fmt.Errorf("model manager is not configured")
	}
	if s.providerManager == nil {
		return 0, fmt.Errorf("provider manager is not configured")
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
