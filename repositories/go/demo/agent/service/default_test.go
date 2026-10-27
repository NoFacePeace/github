package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/event"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/loop"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
)

type loopFunc func(context.Context, loop.Request, func(event.Event)) error

func (f loopFunc) Run(ctx context.Context, req loop.Request, onEvent func(event.Event)) error {
	return f(ctx, req, onEvent)
}

type stubModel struct{}

func (*stubModel) Generate(context.Context, model.Request, func(model.Event)) error {
	return nil
}

func TestDefaultServiceRun(t *testing.T) {
	for _, tc := range []struct {
		name     string
		request  Request
		provider string
		protocol string
		model    string
	}{
		{"defaults", Request{Input: "hello"}, "default", "chat_completions", "default-model"},
		{"overrides", Request{Input: "hello", Provider: "other", Protocol: "other_protocol", Model: "other-model"}, "other", "other_protocol", "other-model"},
		{"partial override", Request{Input: "hello", Model: "other-model"}, "default", "chat_completions", "other-model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			selected := &stubModel{}
			factoryCalled := false
			providers := map[string]ProviderEntry{
				tc.provider: {Factories: map[string]ModelFactory{
					tc.protocol: func(name string) model.Model {
						factoryCalled = true
						if name != tc.model {
							t.Fatalf("model = %q, want %q", name, tc.model)
						}
						return selected
					},
				}},
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			failure := errors.New("loop failure")
			agentLoop := loopFunc(func(gotCtx context.Context, req loop.Request, onEvent func(event.Event)) error {
				wantMessages := []model.Message{{Role: model.RoleUser, Content: "hello"}}
				if gotCtx != ctx || req.Model != selected || req.MaxSteps != 3 || !reflect.DeepEqual(req.Messages, wantMessages) {
					t.Fatalf("unexpected loop request: %+v", req)
				}
				onEvent(event.TextDeltaEvent{Text: "reply"})
				return failure
			})
			svc := NewDefault(Config{
				Defaults: Defaults{Provider: "default", Protocol: "chat_completions", Model: "default-model"},
				Loop:     LoopConfig{MaxSteps: 3},
			}, providers, agentLoop, nil)
			var received []event.Event
			err := svc.Run(ctx, tc.request, func(e event.Event) { received = append(received, e) })
			if !factoryCalled || !errors.Is(err, failure) || !reflect.DeepEqual(received, []event.Event{event.TextDeltaEvent{Text: "reply"}}) {
				t.Fatalf("factory called = %v, error = %v, events = %+v", factoryCalled, err, received)
			}
		})
	}
}

func TestDefaultServiceSelectionErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request Request
		want    string
	}{
		{"unknown provider", Request{Provider: "missing"}, "unknown provider"},
		{"unsupported protocol", Request{Protocol: "missing"}, "unsupported protocol"},
		{"missing model", Request{}, "model is required"},
		{"unsupported session", Request{Model: "model-a", SessionID: "session-a"}, "sessions are not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewDefault(Config{Defaults: Defaults{Provider: "default", Protocol: "chat_completions"}}, map[string]ProviderEntry{
				"default": {Factories: map[string]ModelFactory{
					"chat_completions": func(string) model.Model {
						t.Fatal("factory should not be called")
						return nil
					},
				}},
			}, loopFunc(func(context.Context, loop.Request, func(event.Event)) error {
				t.Fatal("loop should not be called")
				return nil
			}), nil)
			err := svc.Run(context.Background(), tc.request, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}
