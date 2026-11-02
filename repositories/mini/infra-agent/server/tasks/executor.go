package tasks

import (
	"context"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// toolExecutor 协调本次执行的工具恢复、审批、执行及结果保存。
type toolExecutor struct {
	store  store
	taskID string
	runID  string
	lease  *lease
}

// Execute 批量处理工具调用，返回执行结果或等待状态。
func (e *toolExecutor) Execute(ctx context.Context, calls []llm.ToolCall) ([]tools.Result, error) {
	// TODO: 检查执行记录和审批，查找并执行工具，保存开始及结果记录。
	return nil, nil
}
