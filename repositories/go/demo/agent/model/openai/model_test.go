package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/loop"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type echoTool struct{}

func (echoTool) Definition() tool.Definition {
	return tool.Definition{
		Name: "echo", Description: "Echo text",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`),
	}
}

func (echoTool) Execute(ctx context.Context, args json.RawMessage) (tool.Result, error) {
	var input struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return tool.Result{}, err
	}
	return tool.Result{Content: input.Text}, nil
}

func TestGenerateWithLoop(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req completionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Model != "test-model" || !req.Stream || len(req.Tools) != 1 || req.Tools[0].Function.Name != "echo" {
			t.Errorf("unexpected payload: %+v", req)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if requests == 1 {
			if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
				t.Errorf("unexpected initial messages: %+v", req.Messages)
			}
			io.WriteString(w, "data: "+`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"echo","arguments":"{\"text\":"}}]}}]}`+"\n\n")
			io.WriteString(w, "data: "+`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"hello\"}"}}]},"finish_reason":"tool_calls"}]}`+"\n\n")
		} else {
			if len(req.Messages) != 4 {
				t.Errorf("unexpected history: %+v", req.Messages)
				return
			}
			assistant, result := req.Messages[2], req.Messages[3]
			if assistant.Role != "assistant" || len(assistant.ToolCalls) != 1 || assistant.ToolCalls[0].Function.Arguments != `{"text":"hello"}` {
				t.Errorf("unexpected assistant message: %+v", assistant)
			}
			if result.Role != "tool" || result.ToolCallID != "call_1" || result.Content != "hello" {
				t.Errorf("unexpected tool result: %+v", result)
			}
			io.WriteString(w, "data: "+`{"choices":[{"index":0,"delta":{"content":"Hello"}}]}`+"\n\n")
			io.WriteString(w, "data: "+`{"choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}`+"\n\n")
		}
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	m := New(Config{APIKey: "test-key", Model: "test-model", BaseURL: server.URL + "/v1/"})
	var text strings.Builder
	var reason loop.StopReason
	err := (&loop.DefaultLoop{}).Run(context.Background(), loop.Request{
		Model: m,
		Messages: []model.Message{
			{Role: model.RoleSystem, Content: "Use tools"},
			{Role: model.RoleUser, Content: "Echo hello"},
		},
		Tools: []tool.Tool{echoTool{}},
	}, func(event loop.Event) {
		switch e := event.(type) {
		case loop.TextDeltaEvent:
			text.WriteString(e.Text)
		case loop.CompletedEvent:
			reason = e.Reason
		}
	})
	if err != nil || requests != 2 || text.String() != "Hello!" || reason != loop.StopCompleted {
		t.Fatalf("error = %v, requests = %d, text = %q, reason = %q", err, requests, text.String(), reason)
	}
}

func TestInterleavedToolCalls(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"choices":[{"delta":{"tool_calls":[{"index":1,"id":"b","function":{"name":"second","arguments":"{"}},{"index":0,"id":"a","function":{"name":"first","arguments":"{}"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":1,"function":{"arguments":"}"}}]},"finish_reason":"tool_calls"}]}`,
		`data: {"choices":[]}`,
		`data: [DONE]`,
	}, "\r\n\r\n") + "\r\n\r\n"
	var events []model.Event
	if err := readStream(strings.NewReader(stream), func(e model.Event) { events = append(events, e) }); err != nil {
		t.Fatal(err)
	}
	want := []model.Event{model.CompletedEvent{Message: model.Message{
		Role: model.RoleAssistant,
		ToolCalls: []tool.Call{
			{ID: "a", Name: "first", Arguments: json.RawMessage(`{}`)},
			{ID: "b", Name: "second", Arguments: json.RawMessage(`{}`)},
		},
	}}}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %+v, want %+v", events, want)
	}
}

func TestStreamErrors(t *testing.T) {
	cases := []struct {
		name   string
		stream string
	}{
		{"malformed chunk", "data: invalid\n\n"},
		{"truncated stream", "data: {\"choices\":[]}\n\n"},
		{"missing finish reason", "data: [DONE]\n\n"},
		{"token limit", "data: {\"choices\":[{\"finish_reason\":\"length\"}]}\n\ndata: [DONE]\n\n"},
		{"API error", "data: {\"error\":{\"message\":\"service unavailable\"}}\n\n"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := readStream(strings.NewReader(tt.stream), func(event model.Event) {
				if _, ok := event.(model.CompletedEvent); ok {
					t.Error("failed stream emitted completion")
				}
			})
			if err == nil {
				t.Fatal("expected a stream error")
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"message":"invalid API key"}}`)
	}))
	defer server.Close()
	err := New(Config{BaseURL: server.URL}).Generate(context.Background(), model.Request{}, func(model.Event) {
		t.Error("HTTP error emitted an event")
	})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStreamCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := New(Config{BaseURL: server.URL}).Generate(ctx, model.Request{}, func(event model.Event) {
		if _, ok := event.(model.CompletedEvent); ok {
			t.Error("canceled stream emitted completion")
		}
		cancel()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}
