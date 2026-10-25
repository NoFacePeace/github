package loop

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

const defaultMaxSteps = 10

type DefaultLoop struct{}

func (l *DefaultLoop) Run(ctx context.Context, req Request, onEvent func(Event)) error {
	if req.Model == nil {
		return fmt.Errorf("model is required")
	}
	if req.MaxSteps < 0 {
		return fmt.Errorf("max steps must not be negative")
	}
	maxSteps := req.MaxSteps
	if maxSteps == 0 {
		maxSteps = defaultMaxSteps
	}

	tools := make(map[string]tool.Tool, len(req.Tools))
	definitions := make([]tool.Definition, 0, len(req.Tools))
	for _, t := range req.Tools {
		if t == nil {
			return fmt.Errorf("tool must not be nil")
		}
		definition := t.Definition()
		if strings.TrimSpace(definition.Name) == "" {
			return fmt.Errorf("tool name is required")
		}
		if _, exists := tools[definition.Name]; exists {
			return fmt.Errorf("duplicate tool name: %s", definition.Name)
		}
		tools[definition.Name] = t
		definitions = append(definitions, definition)
	}
	emit := func(event Event) {
		if onEvent != nil {
			onEvent(event)
		}
	}
	// 复制消息切片，避免追加消息时覆盖调用方切片的底层数组。
	messages := append([]model.Message(nil), req.Messages...)
	for step := 0; step < maxSteps; step++ {
		var response model.Message
		completed := false
		var eventErr error
		accept := func(event model.Event) {
			if eventErr != nil {
				return
			}
			if completed {
				eventErr = fmt.Errorf("model emitted an event after completion")
				return
			}
			switch e := event.(type) {
			case model.TextDeltaEvent:
				emit(TextDeltaEvent{Text: e.Text})
			case model.CompletedEvent:
				response = e.Message
				completed = true
			}
		}
		err := req.Model.Generate(ctx, model.Request{
			Messages: messages,
			Tools:    definitions,
		}, func(event model.Event) {
			// 同时接受值和指针形式的模型事件。
			switch e := event.(type) {
			case *model.TextDeltaEvent:
				if e != nil {
					accept(*e)
				}
			case *model.CompletedEvent:
				if e != nil {
					accept(*e)
				}
			default:
				accept(event)
			}
		})
		if err != nil {
			return fmt.Errorf("generate model response: %w", err)
		}
		if eventErr != nil {
			return eventErr
		}
		if !completed {
			return fmt.Errorf("model did not emit a completed event")
		}
		response.Role = model.RoleAssistant
		messages = append(messages, response)
		if len(response.ToolCalls) == 0 {
			emit(CompletedEvent{Reason: StopCompleted})
			return nil
		}
		// 仅依据完整回复执行工具，避免流式工具事件导致重复执行。
		for _, call := range response.ToolCalls {
			emit(ToolCallEvent{Call: call})
			var result tool.Result
			if t, exists := tools[call.Name]; exists {
				result, err = t.Execute(ctx, call.Arguments)
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return fmt.Errorf("execute tool %q: %w", call.Name, err)
				}
				if err != nil {
					result = tool.Result{Content: fmt.Sprintf("tool error: %s", err)}
				}
			} else {
				result.Content = fmt.Sprintf("tool error: unknown tool %q", call.Name)
			}
			emit(ToolResultEvent{CallID: call.ID, Result: result})
			messages = append(messages, model.Message{
				Role:       model.RoleTool,
				Content:    result.Content,
				ToolCallID: call.ID,
			})
		}
	}
	emit(CompletedEvent{Reason: StopMaxSteps})
	return nil
}
