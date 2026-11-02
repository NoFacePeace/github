// Package agentloop 提供模型与工具的执行循环。
package agentloop

import (
	"context"
	"fmt"
	"math"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// Input 定义执行循环的输入。
type Input struct {
	Messages []llm.Message
	Tools    []llm.ToolDefinition
	// MaxSteps 限制本次执行的模型调用轮数。
	MaxSteps int
	// TokenBudget 限制每次模型请求的输入上下文大小。
	TokenBudget int
	// MaxOutputTokens 限制每次模型请求的输出 token 数量。
	MaxOutputTokens int
}

// Runner 定义执行循环的能力。
type Runner interface {
	Run(ctx context.Context, input Input, model llm.Model, executor ToolExecutor, recorder Recorder) error
}

// loop 实现模型与工具的执行循环。
type loop struct {
	config Config
}

// New 创建执行循环实例。
func New(config Config) Runner {
	if config.TrimTriggerPercent <= 0 || config.TrimTriggerPercent > 100 {
		config.TrimTriggerPercent = defaultTrimTriggerPercent
	}
	if config.TrimRetentionPercent <= 0 || config.TrimRetentionPercent > 100 {
		config.TrimRetentionPercent = defaultTrimRetentionPercent
	}
	return &loop{config: config}
}

// Run 执行模型与工具循环。
func (l *loop) Run(ctx context.Context, input Input, model llm.Model, executor ToolExecutor, recorder Recorder) error {
	if input.MaxSteps <= 0 {
		return fmt.Errorf("max steps must be positive")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// 1. 处理已有消息中未完成的工具调用，追加结果；需要审批或输入时记录暂停并返回。
	messages := append([]llm.Message(nil), input.Messages...)
	if calls := pendingToolCalls(messages); len(calls) > 0 {
		updated, paused, err := executeTools(ctx, messages, calls, executor, recorder)
		if err != nil {
			return fmt.Errorf("execute pending tools: %w", err)
		}
		messages = updated
		if paused {
			return nil
		}
	}

	// 2. 进入主循环，检查取消信号、最大步数和上下文预算，必要时压缩上下文。
	lastActualPromptTokens := 0
	lastSentPromptEstimate := 0
	promptTokenScale := 1.0
	for step := 0; step < input.MaxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		promptEstimate := estimateRequestTokens(messages, input.Tools)
		predictedPromptTokens := predictPromptTokens(lastActualPromptTokens, lastSentPromptEstimate, promptEstimate, promptTokenScale)
		if needsTrim(predictedPromptTokens, input.TokenBudget, l.config.TrimTriggerPercent) {
			result := trimMessages(messages, input.TokenBudget, l.config)
			if err := recorder.Record(ctx, Entry{
				Type: EntryTypeContextTrimmed,
				Data: map[string]any{
					"before_tokens":          result.BeforeTokens,
					"after_tokens":           result.AfterTokens,
					"before_bytes":           result.BeforeBytes,
					"after_bytes":            result.AfterBytes,
					"trigger_threshold":      result.TriggerThreshold,
					"retention_target_bytes": result.RetentionTargetBytes,
					"dropped_blocks":         result.DroppedBlocks,
					"truncated_tool_results": result.TruncatedToolResults,
					"target_reached":         result.AfterBytes <= result.RetentionTargetBytes,
				},
			}); err != nil {
				return fmt.Errorf("record context trim: %w", err)
			}
			trimmedEstimate := estimateRequestTokens(result.Messages, input.Tools)
			trimmedPrediction := predictPromptTokens(lastActualPromptTokens, lastSentPromptEstimate, trimmedEstimate, promptTokenScale)
			if needsTrim(trimmedPrediction, input.TokenBudget, l.config.TrimTriggerPercent) {
				return fmt.Errorf("context remains above trim threshold")
			}
			messages = result.Messages
		}

		// 3. 校验工具调用与结果配对，构造请求并调用模型。
		messages = repairToolPairs(messages)
		request := llm.Request{
			Messages:        messages,
			Tools:           input.Tools,
			MaxOutputTokens: input.MaxOutputTokens,
		}
		if step == input.MaxSteps-1 {
			request.ToolChoice = llm.ToolChoiceNone
		}
		lastSentPromptEstimate = estimateRequestTokens(request.Messages, request.Tools)
		response, err := model.Generate(ctx, request)
		if err != nil {
			return fmt.Errorf("generate model response: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		// 4. 记录模型回复，并追加到消息上下文。
		if response.Usage.PromptTokens > 0 {
			lastActualPromptTokens = response.Usage.PromptTokens
			if lastSentPromptEstimate > 0 {
				promptTokenScale = math.Max(promptTokenScale,
					float64(lastActualPromptTokens)/float64(lastSentPromptEstimate))
			}
		}
		if err := recorder.Record(ctx, Entry{
			Type: EntryTypeAssistantMessage,
			Data: map[string]any{
				"content":    response.Content,
				"tool_calls": response.ToolCalls,
				"usage":      response.Usage,
			},
		}); err != nil {
			return fmt.Errorf("record assistant message: %w", err)
		}
		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		})

		// 5. 有工具调用时交给执行器，追加结果后进入下一轮；需要等待时记录暂停并返回。
		if len(response.ToolCalls) > 0 {
			if step == input.MaxSteps-1 {
				return fmt.Errorf("max steps reached without a final response")
			}
			updated, paused, err := executeTools(ctx, messages, response.ToolCalls, executor, recorder)
			if err != nil {
				return fmt.Errorf("execute tools: %w", err)
			}
			messages = updated
			if paused {
				return nil
			}
			continue
		}

		// 6. 没有工具调用时记录完成并退出；故障或取消时返回错误。
		if err := recorder.Record(ctx, Entry{
			Type: EntryTypeCompleted,
			Data: map[string]any{},
		}); err != nil {
			return fmt.Errorf("record completion: %w", err)
		}
		return nil
	}

	return nil
}

// executeTools 统一处理恢复和新生成的工具调用，工具开始与结果记录由执行器负责。
func executeTools(ctx context.Context, messages []llm.Message, calls []llm.ToolCall, executor ToolExecutor, recorder Recorder) ([]llm.Message, bool, error) {
	if err := ctx.Err(); err != nil {
		return messages, false, err
	}
	results, err := executor.Execute(ctx, calls)
	if err != nil {
		return messages, false, err
	}
	if err := ctx.Err(); err != nil {
		return messages, false, err
	}
	expected := make(map[string]bool, len(calls))
	for _, call := range calls {
		expected[call.ID] = true
	}
	byID := make(map[string]tools.Result, len(results))
	var waiting []map[string]any
	for _, result := range results {
		if !expected[result.ToolCallID] {
			return messages, false, fmt.Errorf("unexpected tool result for call %s", result.ToolCallID)
		}
		if _, exists := byID[result.ToolCallID]; exists {
			return messages, false, fmt.Errorf("duplicate tool result for call %s", result.ToolCallID)
		}
		switch result.Status {
		case tools.StatusCompleted:
		case tools.StatusWaitingForApproval, tools.StatusWaitingForInput:
			waiting = append(waiting, map[string]any{
				"tool_call_id": result.ToolCallID,
				"reason":       string(result.Status),
			})
		default:
			return messages, false, fmt.Errorf("unknown tool status %q for call %s", result.Status, result.ToolCallID)
		}
		byID[result.ToolCallID] = result
	}
	for _, call := range calls {
		result, exists := byID[call.ID]
		if !exists {
			if len(waiting) == 0 {
				return messages, false, fmt.Errorf("missing tool result for call %s", call.ID)
			}
			// 执行器因等待而提前停止时，后续调用保留，恢复后再处理。
			continue
		}
		if result.Status == tools.StatusCompleted {
			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    result.Content,
				ToolCallID: call.ID,
			})
		}
	}
	if len(waiting) > 0 {
		if err := recorder.Record(ctx, Entry{
			Type: EntryTypePaused,
			Data: map[string]any{"waiting": waiting},
		}); err != nil {
			return messages, false, fmt.Errorf("record execution pause: %w", err)
		}
		return messages, true, nil
	}
	return messages, false, nil
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
