package agentloop

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

type modelFunc func(context.Context, llm.Request) (llm.Response, error)

func (f modelFunc) Generate(ctx context.Context, request llm.Request) (llm.Response, error) {
	return f(ctx, request)
}

type executorFunc func(context.Context, []llm.ToolCall) ([]tools.Result, error)

func (f executorFunc) Execute(ctx context.Context, calls []llm.ToolCall) ([]tools.Result, error) {
	return f(ctx, calls)
}

type recorderFunc func(context.Context, Entry) error

func (f recorderFunc) Record(ctx context.Context, entry Entry) error {
	return f(ctx, entry)
}

func TestRunToolThenFinalResponse(t *testing.T) {
	call := llm.ToolCall{ID: "call_1", Name: "read"}
	var entries []Entry
	recorder := recorderFunc(func(_ context.Context, entry Entry) error {
		entries = append(entries, entry)
		return nil
	})
	count := 0
	model := modelFunc(func(_ context.Context, request llm.Request) (llm.Response, error) {
		count++
		if request.MaxOutputTokens != 100 {
			t.Fatal("output limit was not passed to model")
		}
		if count == 1 {
			return llm.Response{ToolCalls: []llm.ToolCall{call}, Usage: llm.Usage{PromptTokens: 10}}, nil
		}
		if request.ToolChoice != llm.ToolChoiceNone {
			t.Fatal("last step must disable tools")
		}
		want := []llm.Message{
			{Role: llm.RoleUser, Content: "goal"},
			{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{call}},
			{Role: llm.RoleTool, ToolCallID: call.ID, Content: "file content"},
		}
		if !reflect.DeepEqual(request.Messages, want) {
			t.Fatalf("tool exchange missing from next request: %#v", request.Messages)
		}
		return llm.Response{Content: "done"}, nil
	})
	executor := executorFunc(func(_ context.Context, calls []llm.ToolCall) ([]tools.Result, error) {
		if len(entries) != 1 || entries[0].Type != EntryTypeAssistantMessage {
			t.Fatal("assistant tool calls must be recorded before execution")
		}
		return []tools.Result{{ToolCallID: calls[0].ID, Status: tools.StatusCompleted, Content: "file content"}}, nil
	})
	err := New(Config{}).Run(context.Background(), Input{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "goal"}}, MaxSteps: 2, MaxOutputTokens: 100,
	}, model, executor, recorder)
	if err != nil || count != 2 || len(entries) != 3 || entries[2].Type != EntryTypeCompleted {
		t.Fatalf("err=%v, calls=%d, entries=%#v", err, count, entries)
	}
	if entries[0].Data["usage"].(llm.Usage).PromptTokens != 10 {
		t.Fatal("reported usage was not recorded")
	}
}

func TestRunPausesWithoutAnotherModelCall(t *testing.T) {
	count := 0
	var entries []Entry
	err := New(Config{}).Run(context.Background(), Input{MaxSteps: 3},
		modelFunc(func(context.Context, llm.Request) (llm.Response, error) {
			count++
			return llm.Response{ToolCalls: []llm.ToolCall{{ID: "a"}, {ID: "b"}, {ID: "c"}}}, nil
		}),
		executorFunc(func(context.Context, []llm.ToolCall) ([]tools.Result, error) {
			return []tools.Result{
				{ToolCallID: "a", Status: tools.StatusCompleted, Content: "ok"},
				{ToolCallID: "b", Status: tools.StatusWaitingForApproval},
			}, nil
		}),
		recorderFunc(func(_ context.Context, entry Entry) error {
			entries = append(entries, entry)
			return nil
		}))
	if err != nil || count != 1 || len(entries) != 2 || entries[1].Type != EntryTypePaused {
		t.Fatalf("err=%v, calls=%d, entries=%#v", err, count, entries)
	}
}

func TestRunStopsOnRecordingFailure(t *testing.T) {
	want := errors.New("storage unavailable")
	err := New(Config{}).Run(context.Background(), Input{MaxSteps: 2},
		modelFunc(func(context.Context, llm.Request) (llm.Response, error) {
			return llm.Response{ToolCalls: []llm.ToolCall{{ID: "a"}}}, nil
		}),
		executorFunc(func(context.Context, []llm.ToolCall) ([]tools.Result, error) {
			t.Fatal("tools executed without a durable assistant message")
			return nil, nil
		}),
		recorderFunc(func(context.Context, Entry) error { return want }))
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped recording error", err)
	}
}

func TestRunRejectsToolCallsAtStepLimit(t *testing.T) {
	var entries []EntryType
	err := New(Config{}).Run(context.Background(), Input{MaxSteps: 1},
		modelFunc(func(context.Context, llm.Request) (llm.Response, error) {
			return llm.Response{ToolCalls: []llm.ToolCall{{ID: "a"}}}, nil
		}),
		executorFunc(func(context.Context, []llm.ToolCall) ([]tools.Result, error) {
			t.Fatal("tools executed after step limit")
			return nil, nil
		}),
		recorderFunc(func(_ context.Context, entry Entry) error {
			entries = append(entries, entry.Type)
			return nil
		}))
	if err == nil || !reflect.DeepEqual(entries, []EntryType{EntryTypeAssistantMessage}) {
		t.Fatalf("err=%v, entries=%v", err, entries)
	}
}

func TestRunRejectsMissingToolResult(t *testing.T) {
	count := 0
	err := New(Config{}).Run(context.Background(), Input{MaxSteps: 3},
		modelFunc(func(context.Context, llm.Request) (llm.Response, error) {
			count++
			return llm.Response{ToolCalls: []llm.ToolCall{{ID: "a"}}}, nil
		}),
		executorFunc(func(context.Context, []llm.ToolCall) ([]tools.Result, error) { return nil, nil }),
		recorderFunc(func(context.Context, Entry) error { return nil }))
	if err == nil || count != 1 {
		t.Fatalf("err=%v, model calls=%d", err, count)
	}
}

func TestRunIgnoresModelResponseAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := New(Config{}).Run(ctx, Input{MaxSteps: 1},
		modelFunc(func(context.Context, llm.Request) (llm.Response, error) {
			cancel()
			return llm.Response{Content: "late response"}, nil
		}), nil,
		recorderFunc(func(context.Context, Entry) error {
			t.Fatal("response recorded after cancellation")
			return nil
		}))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}
