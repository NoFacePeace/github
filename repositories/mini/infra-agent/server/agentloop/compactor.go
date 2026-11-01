package agentloop

import (
	"encoding/json"
	"fmt"
	"strings"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// TrimResult 表示裁剪后的消息和统计信息。
type TrimResult struct {
	Messages             []llm.Message
	BeforeTokens         int
	AfterTokens          int
	BeforeBytes          int
	AfterBytes           int
	TriggerThreshold     int
	RetentionTargetBytes int
	DroppedBlocks        int
	TruncatedToolResults int
}

// needsTrim 判断预测输入用量是否超过裁剪阈值。
func needsTrim(predictedTokens, tokenBudget, triggerPercent int) bool {
	return predictedTokens > tokenBudget*triggerPercent/100
}

// trimMessages 裁剪消息并保留完整的工具调用与结果配对，由调用方判断是否需要裁剪。
func trimMessages(messages []llm.Message, tokenBudget int) TrimResult {
	messages = dedupResults(messages)
	// TODO: 实现消息裁剪和统计。
	return TrimResult{Messages: messages}
}

// dedupResults 对消息中的重复工具结果进行去重。
func dedupResults(messages []llm.Message) []llm.Message {
	latest := make(map[string]string)
	duplicates := make(map[string]string)
	for _, message := range messages {
		if message.Role != llm.RoleAssistant {
			continue
		}
		for _, call := range message.ToolCalls {
			if call.Name == tools.NameAskUser || call.Name == tools.NameGetSkillFile {
				continue
			}
			key := call.Name + "|" + toolInputKey(call.Arguments)
			if previousID, exists := latest[key]; exists {
				duplicates[previousID] = call.Name
			}
			latest[key] = call.ID
		}
	}
	if len(duplicates) == 0 {
		return messages
	}

	result := append([]llm.Message(nil), messages...)
	for i := range result {
		if result[i].Role != llm.RoleTool {
			continue
		}
		if name, exists := duplicates[result[i].ToolCallID]; exists {
			result[i].Content = fmt.Sprintf("[dedup: %s — same input, see latest result]", name)
		}
	}
	return result
}

// toolInputKey 规范化工具参数，忽略顶层以双下划线开头的内部字段。
func toolInputKey(arguments json.RawMessage) string {
	var input map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(arguments)))
	decoder.UseNumber()
	if !json.Valid(arguments) || decoder.Decode(&input) != nil {
		return string(arguments)
	}
	for key := range input {
		if strings.HasPrefix(key, "__") {
			delete(input, key)
		}
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return string(arguments)
	}
	return string(encoded)
}
