package tasks

import (
	"context"
	"fmt"

	"infra-agent/server/agents"
	"infra-agent/server/workspaces"
)

// Service 定义任务模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateTaskRequest) (Task, error)
}

// CreateTaskRequest 表示创建任务的请求。
type CreateTaskRequest struct {
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
}

// NewTaskService 创建任务服务。
func NewTaskService(workspaceService workspaces.Service, workspaceConfig workspaces.Config, cfg Config) Service {
	return &service{
		config:           cfg,
		manager:          newManager(),
		workspaceService: workspaceService,
		workspaceConfig:  workspaceConfig,
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
	if request.Model == "" {
		request.Model = s.config.DefaultModel
	}
	task := Task{
		TriggerType:        request.TriggerType,
		AgentID:            request.AgentID,
		Model:              request.Model,
		WorkspaceID:        workspaceID,
		WorkspaceDirectory: workspaceDirectory,
	}
	task, err := s.manager.CreateTask(ctx, task)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}
