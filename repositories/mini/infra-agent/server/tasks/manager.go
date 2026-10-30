package tasks

import "context"

// UpdateTaskRequest 表示任务的部分更新参数。
// nil 表示不修改，非 nil 表示使用指针指向的值，包括空字符串。
type UpdateTaskRequest struct {
	Title        *string
	CurrentRunID *string
}

// manager 管理任务的创建和生命周期。
type manager struct {
	storage storage
}

// newManager 创建任务管理器。
func newManager() *manager {
	return &manager{}
}

// Create 创建一个任务。
func (m *manager) Create(ctx context.Context, task Task) (Task, error) {
	if m.storage != nil {
		if err := m.storage.CreateTask(ctx, task); err != nil {
			return Task{}, err
		}
	}

	return task, nil
}

// Update 仅更新请求中非 nil 的任务字段。
func (m *manager) Update(ctx context.Context, taskID string, request UpdateTaskRequest) error {
	// TODO: 将指定任务的部分更新保存到存储层。
	return nil
}
