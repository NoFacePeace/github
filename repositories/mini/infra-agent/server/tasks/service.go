package tasks

import "context"

// TaskService 定义任务模块对外提供的业务接口。
type TaskService interface {
	Create(ctx context.Context) (Task, error)
}

type service struct{}

// NewTaskService 创建任务服务。
func NewTaskService() TaskService {
	return &service{}
}

// Create 创建一个任务。
func (s *service) Create(ctx context.Context) (Task, error) {
	return Task{}, nil
}
