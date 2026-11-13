package compaction

import "context"

// RemoteCompactor 定义使用模型服务专用压缩能力的执行与后台预计算入口。
// 服务端压缩算法由模型服务提供，本接口负责请求组织、结果整理和应用协调。
type RemoteCompactor interface {
	// RunRemoteCompactTask 执行手动远端压缩，可以等待并复用有效的后台候选。
	// 整理压缩项和保留内容并应用到会话后返回，不向调用方直接返回压缩内容。
	RunRemoteCompactTask(ctx context.Context, request Request) error

	// RunInlineRemoteAutoCompactTask 在当前任务内自动执行远端压缩，可传入后台候选。
	// 候选为空时可直接请求服务；候选失效时按请求中的回退策略处理。
	// 返回是否实际应用了压缩结果，跳过应用时返回 false，执行失败时返回错误。
	RunInlineRemoteAutoCompactTask(ctx context.Context, request Request, candidate *BackgroundCandidate) (bool, error)

	// CaptureBackgroundInput 捕获后台压缩所需的历史及关联状态，避免后台读取持续变化的会话。
	// 本方法只准备输入，不发起模型请求，也不替换会话历史。
	CaptureBackgroundInput(ctx context.Context, request Request) (BackgroundInput, error)

	// RunBackgroundCompaction 根据已捕获的输入请求远端压缩，返回待校验的候选结果。
	// 本方法不修改会话历史，也不自行创建后台任务；异步调度由调用方负责。
	RunBackgroundCompaction(ctx context.Context, input BackgroundInput) (BackgroundCandidate, error)
}

// RemoteClient 定义底层模型服务调用，与压缩任务调度和会话状态更新分离。
type RemoteClient interface {
	// Compact 发送压缩请求，收集流式响应并校验请求正常完成且恰好返回一个压缩项。
	// 返回压缩项、响应标识及可选用量，不重建或替换会话历史。
	Compact(ctx context.Context, request RemoteRequest) (RemoteOutput, error)
}
