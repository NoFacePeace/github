package compaction

import "context"

// HistoryBuilder 定义各压缩策略使用的历史整理能力，不直接更新会话状态。
type HistoryBuilder interface {
	// BuildCompactedHistory 将历史前缀、需保留的用户消息和文字摘要组合为新的历史。
	BuildCompactedHistory(prefix []HistoryItem, userMessages []UserMessage, summary string) []HistoryItem

	// BuildCompactionReplacementHistory 为压缩后的历史补充初始指令和必要运行上下文。
	// 返回可交给会话层应用的替换历史，本方法本身不持久化或替换当前历史。
	BuildCompactionReplacementHistory(ctx context.Context, request Request, history []HistoryItem) ([]HistoryItem, error)

	// TruncateRetainedMessagesForRemoteCompaction 按 token 预算裁剪远端压缩需要保留的消息。
	// 应遵守内容顺序和关联约束；预算不足以保留消息时可以返回空切片。
	TruncateRetainedMessagesForRemoteCompaction(items []HistoryItem, tokenBudget int) []HistoryItem
}
