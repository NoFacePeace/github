package anthropic

import (
	"context"
	"encoding/json"
	"fmt"

	ai "github.com/anthropics/anthropic-sdk-go"
)

// 默认 agent loop 最大轮数，防止工具调用无限循环。
const defaultMaxRounds = 10

// AgentLoopParams 是工具调用 agent loop 的输入参数。
type AgentLoopParams struct {
	Model       string
	MaxTokens   int64
	System      string
	Messages    []Message
	Tools       []Tool
	Temperature float64
	// MaxRounds 是安全轮数上限，0 表示默认 10。
	MaxRounds int
}

// AgentLoopResult 是 agent loop 结束后的结果。
type AgentLoopResult struct {
	// Messages 是完整对话（含每轮助手回复与工具结果），使用官方类型。
	Messages []ai.MessageParam
	// Final 是最后一条助手消息。
	Final *ai.Message
	Text  string
}

// AgentLoop 运行 模型→工具调用→工具结果→模型 的循环，直到助手回复不含
// tool_use 块或达到 MaxRounds。每次工具调用会分发给对应 Tool.Call 处理。
func (c *Client) AgentLoop(ctx context.Context, params AgentLoopParams) (*AgentLoopResult, error) {
	registry := NewToolRegistry(params.Tools...)

	messages, err := toSDKMessages(params.Messages)
	if err != nil {
		return nil, err
	}

	maxRounds := params.MaxRounds
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}

	var last *ai.Message
	for round := 1; round <= maxRounds; round++ {
		sdkParams, err := c.buildParams(params.Model, params.MaxTokens, params.System, messages, params.Tools, params.Temperature, nil)
		if err != nil {
			return nil, err
		}
		msg, err := c.sdk.Messages.New(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("agent loop round %d error: [%w]", round, err)
		}
		last = msg

		blocks, err := extractBlocks(msg)
		if err != nil {
			return nil, err
		}

		// 收集本轮工具调用。
		var calls []Block
		for _, b := range blocks {
			if b.Type == BlockToolUse {
				calls = append(calls, b)
			}
		}
		if len(calls) == 0 {
			return &AgentLoopResult{
				Messages: messages,
				Final:    last,
				Text:     extractText(last),
			}, nil
		}

		// 追加助手回复。
		messages = append(messages, msg.ToParam())

		// 执行工具调用，构造工具结果。
		toolResults := make([]ai.ContentBlockParamUnion, 0, len(calls))
		for _, call := range calls {
			tool, ok := registry.Get(call.ToolName)
			result, isErr := dispatch(ctx, tool, ok, call.ToolInput)
			toolResults = append(toolResults, ai.NewToolResultBlock(call.ToolUseID, result, isErr))
		}
		messages = append(messages, ai.NewUserMessage(toolResults...))
	}

	return &AgentLoopResult{
		Messages: messages,
		Final:    last,
		Text:     extractText(last),
	}, fmt.Errorf("agent loop error: [reached max rounds %d]", maxRounds)
}

// dispatch 执行单个工具调用，返回结果文本与是否出错。
func dispatch(ctx context.Context, tool Tool, found bool, input json.RawMessage) (string, bool) {
	if !found {
		return "tool not found", true
	}
	out, err := tool.Call(ctx, input)
	if err != nil {
		return fmt.Sprintf("tool call error: %s", err.Error()), true
	}
	return out, false
}
