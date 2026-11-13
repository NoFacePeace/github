package compaction

import "context"

// TokenBudgetCompactor 定义通过重建上下文窗口管理预算的压缩入口，不请求模型生成摘要。
// 手动与自动入口应复用内部窗口重建逻辑，并沿用压缩的事件与 Hook 生命周期。
type TokenBudgetCompactor interface {
	// RunManualCompactTask 执行手动窗口重建，保留配置指定的内容并应用新的初始上下文。
	// 独立任务的创建与结束由调用方负责，成功返回表示会话已切换到新窗口。
	RunManualCompactTask(ctx context.Context, request Request) error

	// RunInlineAutoCompactTask 在当前任务内自动重建窗口，不创建新的主任务。
	// 成功后由调用方继续原工作，本方法本身不运行普通请求的模型与工具循环。
	RunInlineAutoCompactTask(ctx context.Context, request Request) error
}
