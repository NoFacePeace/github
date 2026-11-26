package executor

import (
	"context"
	"reflect"
	"testing"
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
