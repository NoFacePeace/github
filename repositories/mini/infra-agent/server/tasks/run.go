package tasks

import "context"

// run 表示任务的一次运行。
type run struct{}

// runManager 管理任务运行的生命周期。
type runManager struct{}

// newRunManager 创建运行管理器。
func newRunManager() *runManager {
	return &runManager{}
}

// start 启动一次任务运行。
func (m *runManager) start(ctx context.Context, item run) error {
	return nil
}
