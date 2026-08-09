package tasks

import "context"

// storage 定义任务数据的持久化能力。
type storage interface {
	CreateTask(ctx context.Context, task Task) error
	CreateRun(ctx context.Context, item run) error
}
