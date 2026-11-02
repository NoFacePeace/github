package agentloop

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

const (
	keepRecentToolResults  = 3
	minTruncateResultBytes = 20 * 1024
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
func trimMessages(messages []llm.Message, tokenBudget int, config Config) TrimResult {
	result := TrimResult{
		BeforeTokens:     estimateRequestTokens(messages, nil),
		BeforeBytes:      messagesByteSize(messages),
		TriggerThreshold: tokenBudget * config.TrimTriggerPercent / 100,
	}
	messages = dedupResults(messages)
	messages = dedupPlans(messages)
	messages, truncated := truncateResults(messages)
	result.TruncatedToolResults = truncated
	result.RetentionTargetBytes = messagesByteSize(messages) * config.TrimRetentionPercent / 100
	for messagesByteSize(messages) > result.RetentionTargetBytes {
		start, end, ok := oldestDroppableBlock(messages)
		if !ok {
			break
		}
		messages = append(messages[:start], messages[end:]...)
		result.DroppedBlocks++
	}
	result.Messages = messages
	result.AfterTokens = estimateRequestTokens(messages, nil)
	result.AfterBytes = messagesByteSize(messages)
	return result
}

// messagesByteSize 统计消息文本与工具协议字段的字节数。
func messagesByteSize(messages []llm.Message) int {
	total := 0
	for _, message := range messages {
		total += len(message.Role) + len(message.Content) + len(message.ToolCallID)
		for _, call := range message.ToolCalls {
			total += len(call.ID) + len(call.Name) + len(call.Arguments)
		}
	}
	return total
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

// dedupPlans 只保留最新的计划调用及结果，保留其他调用和消息正文。
func dedupPlans(messages []llm.Message) []llm.Message {
	latestID := ""
	removed := make(map[string]bool)
	for _, message := range messages {
		if message.Role != llm.RoleAssistant {
			continue
		}
		for _, call := range message.ToolCalls {
			if call.Name != tools.NameUpdatePlan {
				continue
			}
			if latestID != "" {
				removed[latestID] = true
			}
			latestID = call.ID
		}
	}
	if len(removed) == 0 {
		return messages
	}

	result := make([]llm.Message, 0, len(messages))
	for _, message := range messages {
		if message.Role == llm.RoleTool && removed[message.ToolCallID] {
			continue
		}
		if message.Role == llm.RoleAssistant && len(message.ToolCalls) > 0 {
			calls := make([]llm.ToolCall, 0, len(message.ToolCalls))
			for _, call := range message.ToolCalls {
				if !removed[call.ID] {
					calls = append(calls, call)
				}
			}
			if len(calls) == 0 && message.Content == "" {
				continue
			}
			message.ToolCalls = calls
		}
		result = append(result, message)
	}
	return result
}

// truncateResults 截短较早的大段工具输出，返回处理后的消息和截短数量。
func truncateResults(messages []llm.Message) ([]llm.Message, int) {
	result := append([]llm.Message(nil), messages...)
	calls := make(map[string]llm.ToolCall)
	var indices []int
	for i, message := range messages {
		if message.Role == llm.RoleAssistant {
			for _, call := range message.ToolCalls {
				calls[call.ID] = call
			}
		}
		if message.Role == llm.RoleTool {
			indices = append(indices, i)
		}
	}
	cutoff := len(indices) - keepRecentToolResults
	if cutoff <= 0 {
		return result, 0
	}

	truncated := 0
	for _, i := range indices[:cutoff] {
		call := calls[result[i].ToolCallID]
		if call.Name == tools.NameAskUser {
			continue
		}
		if call.Name == tools.NameGetSkillFile {
			result[i].Content = skillReloadMarker(call, result[i].Content)
			truncated++
			continue
		}
		if len(result[i].Content) > minTruncateResultBytes {
			result[i].Content = resultPreview(result[i], "old tool result compacted from model context")
			truncated++
		}
	}
	return result, truncated
}

// resultPreview 生成结果替换提示，不声明尚未实现的日志读取工具。
func resultPreview(message llm.Message, reason string) string {
	data, _ := json.Marshal(map[string]any{
		"truncated":      true,
		"tool_call_id":   message.ToolCallID,
		"original_bytes": len(message.Content),
		"reason":         reason,
	})
	return string(data)
}

// skillReloadMarker 提醒模型重新加载指定技能，不再依赖已裁剪的指令。
func skillReloadMarker(call llm.ToolCall, content string) string {
	var params struct {
		SkillID      string `json:"skill_id"`
		WorkspaceID  string `json:"workspace_id"`
		ResourcePath string `json:"resource_path"`
	}
	if json.Unmarshal(call.Arguments, &params) != nil || params.SkillID == "" {
		return resultPreview(llm.Message{ToolCallID: call.ID, Content: content},
			"skill content compacted; inspect the original tool call before retrying")
	}
	digest := sha256.Sum256([]byte(content))
	data, _ := json.Marshal(map[string]any{
		"compacted":      true,
		"tool":           tools.NameGetSkillFile,
		"skill_id":       params.SkillID,
		"workspace_id":   params.WorkspaceID,
		"resource_path":  params.ResourcePath,
		"content_sha256": fmt.Sprintf("%x", digest),
		"action":         "Reload this exact skill_id and workspace_id with get_skill_file before relying on its instructions.",
	})
	return string(data)
}

// repairToolPairs 整理消息中的工具调用与结果配对，不执行工具或修改原始日志。
func repairToolPairs(messages []llm.Message) []llm.Message {
	// TODO: 处理缺少结果的调用和缺少调用的结果。
	return messages
}

// oldestDroppableBlock 查找最早可删除的消息块，范围为 [start, end)，ok 表示是否找到。
func oldestDroppableBlock(messages []llm.Message) (start, end int, ok bool) {
	blocks := buildMessageBlocks(messages)
	latestUser := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llm.RoleUser {
			latestUser = i
			break
		}
	}

	// 优先移除最新用户输入之前的历史。
	for _, block := range blocks {
		if block.end <= latestUser && block.droppable {
			return block.start, block.end, true
		}
	}

	// 当前轮只移除较早的完整工具批次，保留最新几批和纯文本消息。
	recentBatches := 0
	protected := make(map[int]bool)
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		if block.start <= latestUser || !block.toolBatch {
			continue
		}
		if recentBatches < keepRecentToolResults {
			protected[block.start] = true
		}
		recentBatches++
	}
	for _, block := range blocks {
		if block.start > latestUser && block.toolBatch && block.droppable && !protected[block.start] {
			return block.start, block.end, true
		}
	}
	return 0, 0, false
}

// messageBlock 表示连续消息范围及其删除资格。
type messageBlock struct {
	start, end int
	toolBatch  bool
	droppable  bool
}

// buildMessageBlocks 将工具调用和连续结果组成完整块，不删除异常或未完成的配对。
func buildMessageBlocks(messages []llm.Message) []messageBlock {
	resultCounts := make(map[string]int)
	callCounts := make(map[string]int)
	for _, message := range messages {
		if message.Role == llm.RoleTool {
			resultCounts[message.ToolCallID]++
		}
		if message.Role == llm.RoleAssistant {
			for _, call := range message.ToolCalls {
				callCounts[call.ID]++
			}
		}
	}
	var blocks []messageBlock
	for i := 0; i < len(messages); {
		message := messages[i]
		block := messageBlock{start: i, end: i + 1,
			droppable: message.Role == llm.RoleUser || message.Role == llm.RoleAssistant}
		if message.Role == llm.RoleAssistant && len(message.ToolCalls) > 0 {
			block.toolBatch = true
			pending := make(map[string]bool)
			for _, call := range message.ToolCalls {
				pending[call.ID] = true
				if call.ID == "" || callCounts[call.ID] != 1 || resultCounts[call.ID] != 1 ||
					call.Name == tools.NameAskUser || call.Name == tools.NameGetSkillFile {
					block.droppable = false
				}
			}
			for block.end < len(messages) {
				next := messages[block.end]
				if next.Role != llm.RoleTool || !pending[next.ToolCallID] {
					break
				}
				delete(pending, next.ToolCallID)
				block.end++
			}
			if len(pending) > 0 {
				block.droppable = false
			}
		}
		blocks = append(blocks, block)
		i = block.end
	}
	return blocks
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
