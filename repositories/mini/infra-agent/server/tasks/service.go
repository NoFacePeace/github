package tasks

import (
	"context"
	"fmt"

	"infra-agent/server/workspaces"
)

// Service 定义任务模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateTaskRequest) (Task, error)
}

// CreateTaskRequest 表示创建任务的请求。
type CreateTaskRequest struct {
	TriggerType TriggerType
	WorkspaceID string
}

type service struct {
	manager          *manager
	workspaceService workspaces.Service
}

// NewTaskService 创建任务服务。
func NewTaskService(workspaceService workspaces.Service) Service {
	return &service{
		manager:          newManager(),
		workspaceService: workspaceService,
	}
}

// Create 创建一个任务。
func (s *service) Create(ctx context.Context, request CreateTaskRequest) (Task, error) {
	if request.TriggerType == "" {
		request.TriggerType = TriggerTypeManual
	}
	if s.workspaceService != nil {
		if _, err := s.workspaceService.Get(ctx, request.WorkspaceID); err != nil {
			return Task{}, fmt.Errorf("get workspace %q: %w", request.WorkspaceID, err)
		}
	}
	task := Task{TriggerType: request.TriggerType}
	return s.manager.CreateTask(ctx, task)
}
