package anthropic

import (
	"encoding/json"
	"fmt"

	ai "github.com/anthropics/anthropic-sdk-go"
)

// Message 是简化的对话消息。
//
// Role 取 "user" 或 "assistant"。
type Message struct {
	Role   string
	Blocks []Block
}

// Block 是简化的内容块，统一表示文本、工具调用与工具结果。
type Block struct {
	// Type 取 "text"、"tool_use" 或 "tool_result"。
	Type      string
	Text      string
	ToolUseID string
	ToolName  string
	// ToolInput 是 "tool_use" 块的输入参数（原始 JSON）。
	ToolInput json.RawMessage
	// IsError 仅对 "tool_result" 有意义，表示工具执行失败。
	IsError bool
}

// 块类型常量。
const (
	BlockText       = "text"
	BlockToolUse    = "tool_use"
	BlockToolResult = "tool_result"
)

// NewUserMessage 构造一条包含若干文本块的用户消息。
func NewUserMessage(texts ...string) Message {
	blocks := make([]Block, 0, len(texts))
	for _, t := range texts {
		blocks = append(blocks, Block{Type: BlockText, Text: t})
	}
	return Message{Role: "user", Blocks: blocks}
}

// NewAssistantMessage 构造一条包含若干文本块的助手消息。
func NewAssistantMessage(texts ...string) Message {
	blocks := make([]Block, 0, len(texts))
	for _, t := range texts {
		blocks = append(blocks, Block{Type: BlockText, Text: t})
	}
	return Message{Role: "assistant", Blocks: blocks}
}

// NewToolResultMessage 构造一条包含单个工具结果的用户消息。
func NewToolResultMessage(toolUseID, content string, isError bool) Message {
	return Message{
		Role: "user",
		Blocks: []Block{{
			Type:      BlockToolResult,
			ToolUseID: toolUseID,
			Text:      content,
			IsError:   isError,
		}},
	}
}

// toSDKMessage 将简化的 Message 转换为官方 SDK 的 MessageParam。
func toSDKMessage(m Message) (ai.MessageParam, error) {
	role := ai.MessageParamRoleUser
	if m.Role == "assistant" {
		role = ai.MessageParamRoleAssistant
	} else if m.Role != "user" {
		return ai.MessageParam{}, fmt.Errorf("to sdk message error: [unknown role %q]", m.Role)
	}

	userBlocks := make([]ai.ContentBlockParamUnion, 0, len(m.Blocks))
	for _, b := range m.Blocks {
		switch b.Type {
		case BlockText:
			userBlocks = append(userBlocks, ai.NewTextBlock(b.Text))
		case BlockToolUse:
			return ai.MessageParam{}, fmt.Errorf("to sdk message error: [tool_use blocks should come from assistant replies, not be built by hand]")
		case BlockToolResult:
			userBlocks = append(userBlocks, ai.NewToolResultBlock(b.ToolUseID, b.Text, b.IsError))
		default:
			return ai.MessageParam{}, fmt.Errorf("to sdk message error: [unknown block type %q]", b.Type)
		}
	}
	return ai.MessageParam{Role: role, Content: userBlocks}, nil
}

// toSDKMessages 批量转换。
func toSDKMessages(ms []Message) ([]ai.MessageParam, error) {
	out := make([]ai.MessageParam, 0, len(ms))
	for _, m := range ms {
		p, err := toSDKMessage(m)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// extractBlocks 从官方 Message.Content 抽取简化块。assistant 回复只可能包含
// text 与 tool_use 块；tool_result 块只存在于调用方构造的 user 消息里，不会出现
// 在 response 中。
func extractBlocks(msg *ai.Message) ([]Block, error) {
	if msg == nil {
		return nil, nil
	}
	blocks := make([]Block, 0, len(msg.Content))
	for _, c := range msg.Content {
		switch variant := c.AsAny().(type) {
		case ai.TextBlock:
			blocks = append(blocks, Block{Type: BlockText, Text: variant.Text})
		case ai.ToolUseBlock:
			input := variant.Input
			if raw := variant.JSON.Input.Raw(); raw != "" {
				input = json.RawMessage(raw)
			}
			blocks = append(blocks, Block{
				Type:      BlockToolUse,
				ToolUseID: variant.ID,
				ToolName:  variant.Name,
				ToolInput: input,
			})
		}
	}
	return blocks, nil
}

func extractText(msg *ai.Message) string {
	if msg == nil {
		return ""
	}
	var out string
	for _, c := range msg.Content {
		if tb, ok := c.AsAny().(ai.TextBlock); ok {
			out += tb.Text
		}
	}
	return out
}
