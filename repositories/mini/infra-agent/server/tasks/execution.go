package tasks

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"infra-agent/server/tools"
)

// ToolExecution 表示从持久化记录还原的单次工具调用状态。
type ToolExecution struct {
	Started bool
	Result  *tools.Result
}

// execution 保存单次运行的取消能力、结束通知和租约凭证。
type execution struct {
	cancel context.CancelFunc
	done   chan struct{}
	lease  *lease
	// leaseLost 标记续租失败，阻止本执行器继续进行状态收尾。
	leaseLost atomic.Bool
	// cancelRequested 区分用户取消与执行完成后的 Context 清理。
	cancelRequested atomic.Bool
}

// executionRegistry 管理当前进程中的执行登记。
type executionRegistry struct {
	mu     sync.Mutex
	active map[string]*execution
}

// Register 登记指定运行的执行，返回执行记录。
func (r *executionRegistry) Register(runID string, cancel context.CancelFunc) (*execution, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.active[runID]; exists {
		return nil, fmt.Errorf("run %q execution is already registered", runID)
	}
	if r.active == nil {
		r.active = make(map[string]*execution)
	}
	item := &execution{cancel: cancel, done: make(chan struct{})}
	r.active[runID] = item
	return item, nil
}

// Unregister 注销指定执行记录，避免移除同一运行的其他执行。
func (r *executionRegistry) Unregister(runID string, item *execution) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active[runID] == item {
		delete(r.active, runID)
	}
}

// Cancel 通知本进程的执行取消，返回是否找到登记的执行。
func (r *executionRegistry) Cancel(runID string) bool {
	r.mu.Lock()
	item := r.active[runID]
	if item != nil {
		item.cancelRequested.Store(true)
	}
	r.mu.Unlock()
	if item == nil {
		return false
	}
	item.cancel()
	return true
}
