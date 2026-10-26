package chatcompletion

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	protocol "github.com/NoFacePeace/github/repositories/go/demo/agent/protocol/chatcompletion"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type Model struct {
	client *protocol.Client
	name   string
}

var _ model.Model = (*Model)(nil)

func New(client *protocol.Client, name string) *Model {
	return &Model{client: client, name: name}
}

func (m *Model) Generate(ctx context.Context, req model.Request, onEvent func(model.Event)) error {
	payload := protocol.Request{Model: m.name}
	for _, message := range req.Messages {
		converted := protocol.Message{
			Role: string(message.Role), Content: message.Content, ToolCallID: message.ToolCallID,
		}
		for _, call := range message.ToolCalls {
			converted.ToolCalls = append(converted.ToolCalls, protocol.ToolCall{
				ID: call.ID, Type: "function",
				Function: protocol.FunctionCall{Name: call.Name, Arguments: string(call.Arguments)},
			})
		}
		payload.Messages = append(payload.Messages, converted)
	}
	for _, definition := range req.Tools {
		payload.Tools = append(payload.Tools, protocol.Tool{
			Type: "function", Function: protocol.FunctionDefinition{
				Name: definition.Name, Description: definition.Description, Parameters: definition.InputSchema,
			},
		})
	}
	var text strings.Builder
	calls := make(map[int]*protocol.ToolCall)
	finishReason := ""
	if err := m.client.Stream(ctx, payload, func(chunk protocol.Chunk) {
		for _, choice := range chunk.Choices {
			if choice.Index != 0 {
				continue
			}
			if choice.Delta.Content != "" {
				text.WriteString(choice.Delta.Content)
				onEvent(model.TextDeltaEvent{Text: choice.Delta.Content})
			}
			if choice.Delta.Refusal != "" {
				text.WriteString(choice.Delta.Refusal)
				onEvent(model.TextDeltaEvent{Text: choice.Delta.Refusal})
			}
			for _, delta := range choice.Delta.ToolCalls {
				if calls[delta.Index] == nil {
					calls[delta.Index] = &protocol.ToolCall{}
				}
				call := calls[delta.Index]
				call.ID += delta.ID
				call.Function.Name += delta.Function.Name
				call.Function.Arguments += delta.Function.Arguments
			}
			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
		}
	}); err != nil {
		return err
	}
	if finishReason != "stop" && finishReason != "tool_calls" {
		return fmt.Errorf("completion stopped unexpectedly: %q", finishReason)
	}
	message := model.Message{Role: model.RoleAssistant, Content: text.String()}
	indices := make([]int, 0, len(calls))
	for index := range calls {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	for _, index := range indices {
		call := calls[index]
		message.ToolCalls = append(message.ToolCalls, tool.Call{
			ID: call.ID, Name: call.Function.Name, Arguments: json.RawMessage(call.Function.Arguments),
		})
	}
	onEvent(model.CompletedEvent{Message: message})
	return nil
}
