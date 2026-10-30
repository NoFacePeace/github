package tasks

import "context"

// UpdateTaskRequest 表示任务的部分更新参数。
// nil 表示不修改，非 nil 表示使用指针指向的值，包括空字符串。
type UpdateTaskRequest struct {
	Title        *string
	CurrentRunID *string
}

// storage 定义任务数据的持久化能力。
type storage interface {
	CreateTask(ctx context.Context, task Task) error
	UpdateTask(ctx context.Context, taskID string, request UpdateTaskRequest) error
	CreateRun(ctx context.Context, item run) error
}
