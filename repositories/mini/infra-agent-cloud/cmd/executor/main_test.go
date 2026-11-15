package main

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"infra-agent-cloud/executor"
)

func TestControlInputCloseWithFilePipe(t *testing.T) {
	source, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	defer writer.Close()
	input, err := openInput(source)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if err := input.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("input pipe is not pollable: %v", err)
	}
	if _, err := io.WriteString(writer, "{\"executionId\":\"execution-1\"}\n"); err != nil {
		t.Fatal(err)
	}
	control, err := executor.OpenControlInput(context.Background(), input, "execution-1")
	if err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	go func() {
		control.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close failed to unblock a real stdin pipe")
	}
}
