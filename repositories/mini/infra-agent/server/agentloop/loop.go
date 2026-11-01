// Package agentloop 提供模型与工具的执行循环。
package agentloop

import (
	"context"
	"fmt"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// Input 定义执行循环的输入。
type Input struct {
	Messages []llm.Message
	Tools    []tools.Tool
}

// Runner 定义执行循环的能力。
type Runner interface {
	Run(ctx context.Context, input Input, executor ToolExecutor, recorder Recorder) error
}

// loop 实现模型与工具的执行循环。
type loop struct{}

// New 创建执行循环实例。
func New() Runner {
	return &loop{}
}

// Run 执行模型与工具循环。
func (l *loop) Run(ctx context.Context, input Input, executor ToolExecutor, recorder Recorder) error {
	messages := append([]llm.Message(nil), input.Messages...)
	if calls := pendingToolCalls(messages); len(calls) > 0 {
		results, err := executor.Execute(ctx, calls)
		if err != nil {
			return fmt.Errorf("execute pending tools: %w", err)
		}
		var waiting []map[string]any
		for _, result := range results {
			switch result.Status {
			case tools.StatusCompleted:
				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					Content:    result.Content,
					ToolCallID: result.ToolCallID,
				})
			case tools.StatusWaitingForApproval, tools.StatusWaitingForInput:
				waiting = append(waiting, map[string]any{
					"tool_call_id": result.ToolCallID,
					"reason":       string(result.Status),
				})
			default:
				return fmt.Errorf("unknown tool status %q for call %s", result.Status, result.ToolCallID)
			}
		}
		if len(waiting) > 0 {
			if err := recorder.Record(ctx, Entry{
				Type: EntryTypeExecutionPaused,
				Data: map[string]any{"waiting": waiting},
			}); err != nil {
				return fmt.Errorf("record execution pause: %w", err)
			}
			return nil
		}
	}
	// TODO: 使用更新后的消息继续模型调用与工具执行循环。
	return nil
}

// pendingToolCalls 查找最后一条模型消息中缺少后续结果的工具调用。
func pendingToolCalls(messages []llm.Message) []llm.ToolCall {
	lastAssistant := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llm.RoleAssistant {
			lastAssistant = i
			break
		}
	}
	if lastAssistant == -1 {
		return nil
	}

	results := make(map[string]struct{})
	for _, message := range messages[lastAssistant+1:] {
		if message.Role == llm.RoleTool {
			results[message.ToolCallID] = struct{}{}
		}
	}

	var pending []llm.ToolCall
	for _, call := range messages[lastAssistant].ToolCalls {
		if _, exists := results[call.ID]; !exists {
			pending = append(pending, call)
		}
	}
	return pending
}
