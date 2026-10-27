package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/event"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/service"
)

type serviceFunc func(context.Context, service.Request, func(event.Event)) error

func (f serviceFunc) Run(ctx context.Context, req service.Request, onEvent func(event.Event)) error {
	return f(ctx, req, onEvent)
}

func TestInteractiveContinuesAfterFailure(t *testing.T) {
	var inputs []string
	svc := serviceFunc(func(ctx context.Context, req service.Request, emit func(event.Event)) error {
		inputs = append(inputs, req.Input)
		if req.Model != "chosen-model" {
			t.Fatalf("unexpected model: %q", req.Model)
		}
		if req.Input == "bad" {
			return errors.New("request failed")
		}
		emit(event.TextDeltaEvent{Text: "reply"})
		return nil
	})
	var output, errorOutput bytes.Buffer
	err := interactive(context.Background(), svc, service.Request{Model: "chosen-model"},
		strings.NewReader("\nbad\nhello\nquit\nignored\n"), &output, &errorOutput)
	if err != nil || strings.Join(inputs, ",") != "bad,hello" {
		t.Fatalf("error = %v, inputs = %v", err, inputs)
	}
	if !strings.Contains(output.String(), "reply") || !strings.Contains(errorOutput.String(), "request failed") {
		t.Fatalf("output = %q, errors = %q", output.String(), errorOutput.String())
	}
}

func TestInteractiveCancellationWhileWaiting(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- interactive(ctx, serviceFunc(func(context.Context, service.Request, func(event.Event)) error {
			t.Error("service should not run without input")
			return nil
		}), service.Request{}, reader, io.Discard, io.Discard)
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("interactive loop did not stop on cancellation")
	}
}

func TestRunWithConfiguration(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("unexpected endpoint or authentication")
		}
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Model != "selected-model" || len(req.Messages) != 1 {
			t.Errorf("unexpected model or history: %+v", req)
		}
		requests <- req.Model
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello!\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := fmt.Sprintf(`service:
  defaults:
    provider: example
    protocol: chat_completions
    model: default-model
  loop:
    max_steps: 3
providers:
  example:
    base_url: %q
    api_key: test-key
    protocols: [chat_completions]
`, server.URL+"/v1")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	var output, errorOutput bytes.Buffer
	err := run(context.Background(), []string{"-config", path, "-model", "selected-model"},
		strings.NewReader("hello\nagain\n"), &output, &errorOutput)
	if err != nil || len(requests) != 2 || strings.Count(output.String(), "Hello!") != 2 || errorOutput.Len() != 0 {
		t.Fatalf("error = %v, requests = %d, output = %q, errors = %q", err, len(requests), output.String(), errorOutput.String())
	}
}
