package executor

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestConsumeSteerCommandsReusesReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	commands := make(chan SteerCommand, 3)
	commands <- SteerCommand{CommandID: "first", Prompt: "first"}
	commands <- SteerCommand{CommandID: "first", Prompt: "duplicate"}
	commands <- SteerCommand{CommandID: "last", Prompt: "last"}

	var prompts []string
	steer := func(_ context.Context, command SteerCommand) (map[string]any, error) {
		prompts = append(prompts, command.Prompt)
		return map[string]any{"consumed": true}, nil
	}

	var receipts []map[string]any
	publish := func(kind EventType, payload map[string]any) error {
		if kind != EventTypeSteerResult {
			t.Fatalf("unexpected event type %q", kind)
		}
		receipts = append(receipts, payload)
		if len(receipts) == 3 {
			cancel()
		}
		return nil
	}
	// 保持命令流打开，验证取消可以结束命令循环。
	if err := consumeSteerCommands(ctx, commands, steer, publish); err != nil {
		t.Fatal(err)
	}
	if want := []string{"first", "last"}; !reflect.DeepEqual(prompts, want) {
		t.Fatalf("steer prompts = %v, want %v", prompts, want)
	}
	want := []map[string]any{
		{"commandId": "first", "consumed": true},
		{"commandId": "first", "consumed": true},
		{"commandId": "last", "consumed": true},
	}
	if !reflect.DeepEqual(receipts, want) {
		t.Fatalf("steer receipts = %v, want %v", receipts, want)
	}
}

type eofSteerEngine struct{ Codex }

func (*eofSteerEngine) StartSteer(_ context.Context, spec *Spec, stdin io.WriteCloser, stdout io.ReadCloser) (SteerSession, error) {
	if _, err := fmt.Fprintln(stdin, spec.Prompt); err != nil {
		return SteerSession{}, err
	}
	output, writer := io.Pipe()
	go func() {
		err := readEngineOutput(stdout, func(line string) error {
			if line == "{\"type\":\"turn.completed\"}\n" {
				if err := stdin.Close(); err != nil {
					return err
				}
			}
			_, err := io.WriteString(writer, line)
			return err
		})
		_ = writer.CloseWithError(err)
	}()
	return SteerSession{
		Stdout: output,
		Steer: func(context.Context, SteerCommand) (map[string]any, error) {
			return nil, nil
		},
	}, nil
}

func TestRunEngineProcessAllowsSteerToCloseInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := EngineCommand{Executable: "/bin/sh", Args: []string{"-c", `
IFS= read -r initial && [ "$initial" = "initial" ] || exit 1
printf '%s\n' '{"type":"turn.completed"}'
if IFS= read -r leftover; then exit 2; fi
`}}
	spec := &Spec{SessionID: "test", SteerPromptMode: SteerPromptModeSteer, Prompt: "initial", Paths: Paths{Session: t.TempDir()}}
	result, err := runEngineProcess(ctx, &eofSteerEngine{}, command, spec, nil, func(EventType, map[string]any) error { return nil })
	if err != nil || result.ExitCode == nil || *result.ExitCode != 0 {
		t.Fatalf("runEngineProcess() = (%+v, %v)", result, err)
	}
}
