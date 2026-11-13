package compaction

import "context"

// LocalCompactor 定义由客户端组织模型摘要请求并重建历史的压缩入口。
// 模型仍可由远端服务提供；手动与自动入口应复用内部摘要和结果应用逻辑。
type LocalCompactor interface {
	// RunCompactTask 执行手动摘要压缩，生成文字摘要、组合保留内容并应用到会话。
	// 调用方负责独立任务的创建和收尾，本方法返回压缩成功或错误。
	RunCompactTask(ctx context.Context, request Request) error

	// RunInlineAutoCompactTask 在当前普通任务内执行摘要压缩，不创建新的主任务。
	// 成功后由调用方继续原工作，取消或失败时返回错误。
	RunInlineAutoCompactTask(ctx context.Context, request Request) error
}
