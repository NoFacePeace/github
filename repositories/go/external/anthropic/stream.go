package anthropic

import (
	"context"
	"fmt"

	ai "github.com/anthropics/anthropic-sdk-go"
)

// 流式事件类型常量。
const (
	EventMessageStart     = "message_start"
	EventMessageDelta     = "message_delta"
	EventTextDelta        = "text_delta"
	EventInputJSONDelta   = "input_json_delta"
)

// StreamEvent 是流式回调收到的高层事件视图。
type StreamEvent struct {
	Type       string
	Text       string // EventTextDelta 时的文本片段
	ToolUseID  string // EventInputJSONDelta 时关联的工具调用（按出现序回填）
	InputDelta string // EventInputJSONDelta 时的部分 JSON 片段
	StopReason string // EventMessageDelta 时的停止原因
}

// StreamCallback 在每个流事件上被调用，可为 nil。
type StreamCallback func(StreamEvent)

// StreamParams 是流式对话的输入参数，字段含义同 CompleteParams。
type StreamParams struct {
	Model       string
	MaxTokens   int64
	System      string
	Messages    []Message
	Tools       []Tool
	Temperature float64
	StopSeqs    []string
}

// StreamResult 是流式完成后的累积结果。
type StreamResult struct {
	Message *ai.Message
	Text    string
	Blocks  []Block
}

// Stream 发起流式 Messages.NewStreaming 调用，逐事件回调 cb（可为 nil），
// 最终返回累积得到的完整消息。
func (c *Client) Stream(ctx context.Context, params StreamParams, cb StreamCallback) (*StreamResult, error) {
	sdkMessages, err := toSDKMessages(params.Messages)
	if err != nil {
		return nil, err
	}
	sdkParams, err := c.buildParams(params.Model, params.MaxTokens, params.System, sdkMessages, params.Tools, params.Temperature, params.StopSeqs)
	if err != nil {
		return nil, err
	}

	stream := c.sdk.Messages.NewStreaming(ctx, sdkParams)

	var acc ai.Message
	// 追踪当前正在累积的 tool_use 块的 ID，用于把 input_json_delta 关联到具体工具调用。
	var openToolUseIDs []string
	// 轮换记录最近一个 tool_use 块，用于在 delta 到达时回填 ID。

	emit := func(ev StreamEvent) {
		if cb != nil {
			cb(ev)
		}
	}

	for stream.Next() {
		union := stream.Current()
		if err := acc.Accumulate(union); err != nil {
			return nil, fmt.Errorf("stream accumulate error: [%w]", err)
		}
		if cb == nil {
			continue
		}
		switch variant := union.AsAny().(type) {
		case ai.MessageStartEvent:
			emit(StreamEvent{Type: EventMessageStart})
		case ai.MessageDeltaEvent:
			emit(StreamEvent{
				Type:       EventMessageDelta,
				StopReason: string(variant.Delta.StopReason),
			})
		case ai.ContentBlockStartEvent:
			if tu, ok := variant.ContentBlock.AsAny().(ai.ToolUseBlock); ok {
				openToolUseIDs = append(openToolUseIDs, tu.ID)
			}
		case ai.ContentBlockDeltaEvent:
			switch delta := variant.Delta.AsAny().(type) {
			case ai.TextDelta:
				emit(StreamEvent{Type: EventTextDelta, Text: delta.Text})
			case ai.InputJSONDelta:
				id := lastOr(openToolUseIDs, "")
				emit(StreamEvent{
					Type:        EventInputJSONDelta,
					ToolUseID:   id,
					InputDelta:  delta.PartialJSON,
				})
			}
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream consume error: [%w]", err)
	}

	blocks, err := extractBlocks(&acc)
	if err != nil {
		return nil, err
	}
	return &StreamResult{
		Message: &acc,
		Text:    extractText(&acc),
		Blocks:  blocks,
	}, nil
}

func lastOr(s []string, def string) string {
	if len(s) == 0 {
		return def
	}
	return s[len(s)-1]
}
