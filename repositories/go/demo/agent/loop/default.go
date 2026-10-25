package loop

import (
	"context"
	"errors"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

const defaultMaxSteps = 10

type DefaultLoop struct{}

func (l *DefaultLoop) Run(ctx context.Context, req Request, onEvent func(Event)) error {
	maxSteps := req.MaxSteps
	if maxSteps == 0 {
		maxSteps = defaultMaxSteps
	}

	tools := make(map[string]tool.Tool, len(req.Tools))
	definitions := make([]tool.Definition, 0, len(req.Tools))
	for _, t := range req.Tools {
		definition := t.Definition()
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
		if err := req.Model.Generate(ctx, model.Request{
			Messages: messages,
			Tools:    definitions,
		}, func(event model.Event) {
			switch e := event.(type) {
			case model.TextDeltaEvent:
				emit(TextDeltaEvent{Text: e.Text})
			case model.CompletedEvent:
				response = e.Message
				completed = true
			}
		}); err != nil {
			return fmt.Errorf("generate model response: %w", err)
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
		for _, call := range response.ToolCalls {
			emit(ToolCallEvent{Call: call})
			var result tool.Result
			if t, exists := tools[call.Name]; exists {
				output, err := t.Execute(ctx, call.Arguments)
				result = output
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
