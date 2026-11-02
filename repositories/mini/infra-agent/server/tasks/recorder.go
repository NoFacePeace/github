package tasks

import (
	"context"

	"infra-agent/server/agentloop"
)

// recorder 将执行循环记录转换为任务侧的日志、事件和状态更新。
type recorder struct {
	store  store
	taskID string
	runID  string
	lease  *lease
}

// Record 保存本次执行的过程记录，并在受保护的写入中校验租约。
func (r *recorder) Record(ctx context.Context, entry agentloop.Entry) error {
	// TODO: 根据记录类型保存日志、事件或更新运行状态。
	return nil
}
