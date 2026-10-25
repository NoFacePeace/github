package loop

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type modelFunc func(context.Context, model.Request, func(model.Event)) error

func (f modelFunc) Generate(ctx context.Context, req model.Request, onEvent func(model.Event)) error {
	return f(ctx, req, onEvent)
}

type testTool struct {
	execute func(context.Context, json.RawMessage) (tool.Result, error)
}

func (testTool) Definition() tool.Definition {
	return tool.Definition{Name: "echo"}
}

func (t testTool) Execute(ctx context.Context, args json.RawMessage) (tool.Result, error) {
	return t.execute(ctx, args)
}

func TestDefaultLoopToolRoundTrip(t *testing.T) {
	call := tool.Call{ID: "call_1", Name: "echo", Arguments: json.RawMessage(`{"text":"hello"}`)}
	backing := make([]model.Message, 4)
	backing[0] = model.Message{Role: model.RoleUser, Content: "say hello"}
	backing[1] = model.Message{Content: "sentinel"}
	initial := backing[:1]
	modelCalls, toolCalls := 0, 0
	var events []Event
	m := modelFunc(func(ctx context.Context, req model.Request, emit func(model.Event)) error {
		modelCalls++
		if len(req.Tools) != 1 || req.Tools[0].Name != "echo" {
			t.Fatalf("unexpected tool definitions: %+v", req.Tools)
		}
		if modelCalls == 1 {
			emit(model.CompletedEvent{Message: model.Message{ToolCalls: []tool.Call{call}}})
			return nil
		}
		want := []model.Message{
			initial[0],
			{Role: model.RoleAssistant, ToolCalls: []tool.Call{call}},
			{Role: model.RoleTool, Content: "hello", ToolCallID: call.ID},
		}
		if !reflect.DeepEqual(req.Messages, want) {
			t.Fatalf("messages = %+v, want %+v", req.Messages, want)
		}
		emit(model.TextDeltaEvent{Text: "done"})
		emit(model.CompletedEvent{Message: model.Message{Content: "done"}})
		return nil
	})
	err := (&DefaultLoop{}).Run(context.Background(), Request{
		Model: m, Messages: initial,
		Tools: []tool.Tool{testTool{execute: func(ctx context.Context, args json.RawMessage) (tool.Result, error) {
			toolCalls++
			if string(args) != string(call.Arguments) {
				t.Fatalf("unexpected arguments: %s", args)
			}
			return tool.Result{Content: "hello"}, nil
		}}},
	}, func(event Event) { events = append(events, event) })
	if err != nil {
		t.Fatal(err)
	}
	if modelCalls != 2 || toolCalls != 1 {
		t.Fatalf("model calls = %d, tool calls = %d", modelCalls, toolCalls)
	}
	wantEvents := []Event{
		ToolCallEvent{Call: call},
		ToolResultEvent{CallID: call.ID, Result: tool.Result{Content: "hello"}},
		TextDeltaEvent{Text: "done"},
		CompletedEvent{Reason: StopCompleted},
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %+v, want %+v", events, wantEvents)
	}
	if backing[1].Content != "sentinel" {
		t.Fatal("input backing array was modified")
	}
}

func TestDefaultLoopToolErrors(t *testing.T) {
	for _, name := range []string{"echo", "missing"} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			m := modelFunc(func(ctx context.Context, req model.Request, emit func(model.Event)) error {
				calls++
				if calls == 1 {
					emit(model.CompletedEvent{Message: model.Message{ToolCalls: []tool.Call{{ID: "call_1", Name: name}}}})
				} else {
					last := req.Messages[len(req.Messages)-1]
					if last.Role != model.RoleTool || last.ToolCallID != "call_1" || !strings.HasPrefix(last.Content, "tool error:") {
						t.Fatalf("unexpected error feedback: %+v", last)
					}
					emit(model.CompletedEvent{})
				}
				return nil
			})
			err := (&DefaultLoop{}).Run(context.Background(), Request{
				Model: m,
				Tools: []tool.Tool{testTool{execute: func(context.Context, json.RawMessage) (tool.Result, error) {
					return tool.Result{}, errors.New("execution failed")
				}}},
			}, nil)
			if err != nil || calls != 2 {
				t.Fatalf("error = %v, calls = %d", err, calls)
			}
		})
	}
}

func TestDefaultLoopStepLimit(t *testing.T) {
	for _, limit := range []int{0, 2} {
		calls := 0
		m := modelFunc(func(ctx context.Context, req model.Request, emit func(model.Event)) error {
			calls++
			emit(model.CompletedEvent{Message: model.Message{ToolCalls: []tool.Call{{ID: "call", Name: "missing"}}}})
			return nil
		})
		var last Event
		err := (&DefaultLoop{}).Run(context.Background(), Request{Model: m, MaxSteps: limit}, func(e Event) { last = e })
		wantCalls := limit
		if wantCalls == 0 {
			wantCalls = 10
		}
		if err != nil || calls != wantCalls || last != (CompletedEvent{Reason: StopMaxSteps}) {
			t.Fatalf("limit = %d: error = %v, calls = %d, last = %+v", limit, err, calls, last)
		}
	}
}

func TestDefaultLoopFailures(t *testing.T) {
	failure := errors.New("model unavailable")
	cases := []struct {
		name string
		req  Request
		want string
	}{
		{name: "model error", req: Request{Model: modelFunc(func(context.Context, model.Request, func(model.Event)) error {
			return failure
		})}, want: "model unavailable"},
		{name: "missing completion", req: Request{Model: modelFunc(func(context.Context, model.Request, func(model.Event)) error {
			return nil
		})}, want: "did not emit a completed event"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := (&DefaultLoop{}).Run(context.Background(), tt.req, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if tt.name == "model error" && !errors.Is(err, failure) {
				t.Fatal("model error was not preserved")
			}
		})
	}
}

func TestDefaultLoopCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	toolCalls := 0
	m := modelFunc(func(ctx context.Context, req model.Request, emit func(model.Event)) error {
		emit(model.CompletedEvent{Message: model.Message{ToolCalls: []tool.Call{{ID: "call", Name: "echo"}}}})
		return nil
	})
	err := (&DefaultLoop{}).Run(ctx, Request{
		Model: m,
		Tools: []tool.Tool{testTool{execute: func(ctx context.Context, _ json.RawMessage) (tool.Result, error) {
			toolCalls++
			return tool.Result{}, ctx.Err()
		}}},
	}, func(event Event) {
		if _, ok := event.(ToolCallEvent); ok {
			cancel()
		}
	})
	defer cancel()
	if !errors.Is(err, context.Canceled) || toolCalls != 1 {
		t.Fatalf("error = %v, tool calls = %d", err, toolCalls)
	}
}

func TestDefaultLoopModelCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancel()
	calls := 0
	m := modelFunc(func(ctx context.Context, req model.Request, emit func(model.Event)) error {
		calls++
		return ctx.Err()
	})
	err := (&DefaultLoop{}).Run(ctx, Request{Model: m}, nil)
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("error = %v, model calls = %d", err, calls)
	}
}
