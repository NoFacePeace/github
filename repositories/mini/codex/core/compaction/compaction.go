// Package compaction 定义上下文压缩入口、模型请求和历史整理接口，暂不包含执行实现。
package compaction

// Request 定义前台压缩输入，后续补充会话运行时、历史、配置、触发原因和阶段等字段。
// 执行入口通过请求中的运行时应用结果，成功返回表示压缩结果已被应用。
type Request struct{}

// BackgroundInput 定义后台压缩使用的输入快照，后续补充历史、指令、工具与窗口信息。
type BackgroundInput struct{}

// BackgroundCandidate 定义后台生成的待应用结果，后续补充压缩输出及历史版本、模型和认证归属。
// 候选生成成功不表示已修改会话，应用前仍须校验它与当前状态是否兼容。
type BackgroundCandidate struct{}

// RemoteRequest 定义发送给模型服务的压缩请求，后续补充上下文、指令、工具定义和压缩控制项。
type RemoteRequest struct{}

// CompactionItem 定义模型服务返回的专用压缩项，加密内容由支持该协议的模型服务解释。
type CompactionItem struct {
	ID               string
	EncryptedContent string
}

// RemoteOutput 定义远端请求完成后收集的结果，不表示结果已经应用到会话。
type RemoteOutput struct {
	CompactionOutput CompactionItem
	ResponseID       string
	TokenUsage       *TokenUsage
}

// TokenUsage 定义压缩模型请求的用量，具体输入、输出等统计字段后续补充。
type TokenUsage struct{}

// HistoryItem 定义带有必要元数据的历史项，具体消息、工具调用和检查点字段后续补充。
type HistoryItem struct{}

// UserMessage 定义历史重建时保留的用户消息，具体内容与元数据字段后续补充。
type UserMessage struct{}
