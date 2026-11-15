package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestOpenControlInputReturnsBeforeEOF(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	input, writer := io.Pipe()
	defer writer.Close()
	writeDone := make(chan error, 1)
	go func() {
		_, err := io.WriteString(writer, "{\"executionId\":\"execution-1\"}\n")
		writeDone <- err
	}()
	control, err := OpenControlInput(ctx, input, "execution-1")
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	if err := <-writeDone; err != nil {
		t.Fatal(err)
	}
	if control.Secrets.ExecutionID != "execution-1" {
		t.Fatal("first frame was not parsed as secrets")
	}
	if _, err := io.WriteString(writer, "{\"commandId\":\"command-1\",\"executionId\":\"execution-1\",\"prompt\":\"Continue\"}\n"); err != nil {
		t.Fatal(err)
	}
	select {
	case command := <-control.Commands:
		if command.CommandID != "command-1" || command.Prompt != "Continue" {
			t.Fatalf("unexpected command: %+v", command)
		}
	case <-ctx.Done():
		t.Fatal("command was not delivered while stdin remained open")
	}
}

func TestOpenControlInputPreservesBufferedFrames(t *testing.T) {
	var input strings.Builder
	input.WriteString("{\"executionId\":\"execution-1\"}\n")
	for i := range 3 {
		fmt.Fprintf(&input, "{\"commandId\":\"command-%d\",\"executionId\":\"execution-1\",\"prompt\":\"Continue\"}\n", i)
	}
	control, err := OpenControlInput(context.Background(), io.NopCloser(strings.NewReader(input.String())), "execution-1")
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	count := 0
	for command := range control.Commands {
		if command.CommandID != fmt.Sprintf("command-%d", count) {
			t.Fatalf("out-of-order command: %s", command.CommandID)
		}
		count++
	}
	if count != 3 {
		t.Fatalf("received %d commands, expected 3", count)
	}
	if err := <-control.Errors; err != nil {
		t.Fatal(err)
	}
	if cause := context.Cause(control.Context); cause != nil {
		t.Fatalf("normal EOF canceled execution: %v", cause)
	}
}

func TestControlInputErrorsCancelExecution(t *testing.T) {
	cases := map[string]string{
		"malformed":        "{private-content}\n",
		"null":             "null\n",
		"wrong execution":  "{\"commandId\":\"command-1\",\"executionId\":\"execution-2\",\"prompt\":\"private-content\"}\n",
		"missing ID":       "{\"executionId\":\"execution-1\",\"prompt\":\"private-content\"}\n",
		"empty prompt":     "{\"commandId\":\"command-1\",\"executionId\":\"execution-1\",\"prompt\":\"  \"}\n",
		"incomplete frame": "{\"commandId\":\"command-1\"}",
		"oversized frame":  strings.Repeat("x", maxInputBytes+1) + "\n",
	}
	for name, frame := range cases {
		t.Run(name, func(t *testing.T) {
			input := "{\"executionId\":\"execution-1\"}\n" + frame
			control, err := OpenControlInput(context.Background(), io.NopCloser(strings.NewReader(input)), "execution-1")
			if err != nil {
				t.Fatal(err)
			}
			defer control.Close()
			for range control.Commands {
				t.Error("invalid input was delivered as a command")
			}
			err = <-control.Errors
			if err == nil {
				t.Fatal("input error was not reported")
			}
			<-control.Context.Done()
			if !errors.Is(context.Cause(control.Context), err) {
				t.Fatal("input error was not retained as the execution cancellation cause")
			}
			if strings.Contains(err.Error(), "private-content") {
				t.Fatal("input error exposed payload contents")
			}
		})
	}
}

func TestOpenControlInputRejectsInvalidFirstFrame(t *testing.T) {
	for _, input := range []string{
		"", "null\n", "{private-content}\n", "{\"executionId\":\"execution-2\"}\n",
		"{\"executionId\":\"execution-1\"}", strings.Repeat("x", maxInputBytes+1) + "\n",
	} {
		_, err := OpenControlInput(context.Background(), io.NopCloser(strings.NewReader(input)), "execution-1")
		if err == nil {
			t.Fatal("invalid first frame was accepted")
		}
		if strings.Contains(err.Error(), "private-content") {
			t.Fatal("input error exposed payload contents")
		}
	}
}

func TestControlInputCloseUnblocksFullBuffer(t *testing.T) {
	var input strings.Builder
	input.WriteString("{\"executionId\":\"execution-1\"}\n")
	for i := range commandBufferSize + 2 {
		fmt.Fprintf(&input, "{\"commandId\":\"command-%d\",\"executionId\":\"execution-1\",\"prompt\":\"Continue\"}\n", i)
	}
	control, err := OpenControlInput(context.Background(), io.NopCloser(strings.NewReader(input.String())), "execution-1")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for len(control.Commands) < commandBufferSize {
		select {
		case <-deadline.C:
			t.Fatal("command buffer did not fill")
		case <-tick.C:
		}
	}
	closed := make(chan struct{})
	go func() {
		control.Close()
		control.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("Close blocked on command delivery")
	}
}

func TestControlInputCancellationUnblocksRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input, writer := io.Pipe()
	defer writer.Close()
	go func() { _, _ = io.WriteString(writer, "{\"executionId\":\"execution-1\"}\n") }()
	control, err := OpenControlInput(ctx, input, "execution-1")
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	cancel()
	select {
	case <-control.done:
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation did not unblock stdin reading")
	}
}

func TestInputCancellationBeforeSecrets(t *testing.T) {
	for _, steer := range []bool{false, true} {
		t.Run(fmt.Sprintf("steer=%t", steer), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			input, writer := io.Pipe()
			defer writer.Close()
			var err error
			if steer {
				_, err = OpenControlInput(ctx, input, "execution-1")
			} else {
				_, err = ReadSecrets(ctx, input)
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expected deadline error, received %v", err)
			}
		})
	}
}

func TestReadSecrets(t *testing.T) {
	secrets, err := ReadSecrets(context.Background(), io.NopCloser(strings.NewReader("{\"executionId\":\"execution-1\"}")))
	if err != nil || secrets.ExecutionID != "execution-1" {
		t.Fatalf("ordinary secrets parsing failed: %v", err)
	}
	for _, input := range []string{"", "null", "{private-content}", strings.Repeat("x", maxInputBytes+1)} {
		if _, err := ReadSecrets(context.Background(), io.NopCloser(strings.NewReader(input))); err == nil {
			t.Fatal("invalid ordinary secrets were accepted")
		}
	}
}
