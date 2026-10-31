package tasks

import (
	"context"
	"sync"
)

// execution 保存单次运行的取消能力、结束通知和租约凭证。
type execution struct {
	cancel context.CancelFunc
	done   chan struct{}
	lease  *lease
}

// executionRegistry 管理当前进程中的执行登记。
type executionRegistry struct {
	mu     sync.Mutex
	active map[string]*execution
}

// Register 登记指定运行的执行，返回执行记录。
func (r *executionRegistry) Register(runID string, cancel context.CancelFunc) (*execution, error) {
	// TODO: 检查重复登记并保存执行记录。
	return nil, nil
}
