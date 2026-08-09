package tasks

import "context"

// manager 管理任务的创建和生命周期。
type manager struct {
	storage    storage
	runManager *runManager
}

// newManager 创建任务管理器。
func newManager() *manager {
	return &manager{runManager: newRunManager()}
}

// CreateTask 创建一个任务。
func (m *manager) CreateTask(ctx context.Context) (Task, error) {
	task := Task{}
	firstRun := task.createRun()

	if m.storage != nil {
		if err := m.storage.CreateTask(ctx, task); err != nil {
			return Task{}, err
		}
		if err := m.storage.CreateRun(ctx, firstRun); err != nil {
			return Task{}, err
		}
	}
	if err := m.runManager.start(ctx, firstRun); err != nil {
		return Task{}, err
	}

	return task, nil
}
