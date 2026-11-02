package tasks

import (
	"context"
	"sync"
	"sync/atomic"
)

// execution 保存单次运行的取消能力、结束通知和租约凭证。
type execution struct {
	cancel context.CancelFunc
	done   chan struct{}
	lease  *lease
	// leaseLost 标记续租失败，阻止本执行器继续进行状态收尾。
	leaseLost atomic.Bool
}

// executionRegistry 管理当前进程中的执行登记。
type executionRegistry struct {
	mu     sync.Mutex
	active map[string]*execution
}

// Register 登记指定运行的执行，返回执行记录。
func (r *executionRegistry) Register(runID string, cancel context.CancelFunc) (*execution, error) {
	// TODO: 检查重复登记并保存执行记录。
	return &execution{cancel: cancel, done: make(chan struct{})}, nil
}

// Unregister 注销指定执行记录，避免移除同一运行的其他执行。
func (r *executionRegistry) Unregister(runID string, item *execution) {
	// TODO: 核对执行记录身份并移除登记。
}
