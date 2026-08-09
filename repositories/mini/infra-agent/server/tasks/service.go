package tasks

import "context"

// Service 定义任务模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context) (Task, error)
}

type service struct {
	manager *manager
}

// NewTaskService 创建任务服务。
func NewTaskService() Service {
	return &service{manager: newManager()}
}

// Create 创建一个任务。
func (s *service) Create(ctx context.Context) (Task, error) {
	return s.manager.CreateTask(ctx)
}
