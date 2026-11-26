package executor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRunDrainsOutputAfterInterrupt(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	script := `#!/bin/sh
trap 'printf "%s\n" "{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"interrupted\"}}"; printf "%s\n" "shutdown" >&2; exit 0' INT
printf '%s\n' '{"type":"thread.started","thread_id":"test"}'
for i in 1 2 3 4 5; do /bin/sleep 1; done
`
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	spec := &Spec{Engine: "codex", SessionID: "test", Paths: Paths{Workspace: dir, Config: dir, Session: dir}}
	var events []Event
	result, err := Run(ctx, spec, nil, nil, func(event Event) error {
		events = append(events, event)
		if event.Type == EventTypeStdout && event.Payload["data"] == "{\"type\":\"thread.started\",\"thread_id\":\"test\"}\n" {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) || result.ExitCode == nil || *result.ExitCode != 0 || result.Text == nil || *result.Text != "interrupted" {
		t.Fatalf("Run() = (%+v, %v)", result, err)
	}
	if len(events) != 5 || events[4].Type != EventTypeStopped {
		t.Fatalf("expected output before stopped, got %+v", events)
	}
	for _, event := range events[2:4] {
		if event.Type == EventTypeStderr && event.Payload["data"] == "shutdown\n" {
			return
		}
	}
	t.Fatalf("missing shutdown output: %+v", events)
}
