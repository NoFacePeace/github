package agentloop

import "infra-agent/server/llm"

// estimateRequestTokens 估算消息和工具定义占用的输入 token 数量。
func estimateRequestTokens(messages []llm.Message, tools []llm.ToolDefinition) int {
	// TODO: 实现请求 token 估算。
	return 0
}

// predictPromptTokens 根据上一轮实际用量、估算增量和修正比例预测本轮输入用量。
func predictPromptTokens(lastActual, lastSentEstimate, currentEstimate int, scale float64) int {
	// TODO: 实现输入 token 用量预测。
	return 0
}
