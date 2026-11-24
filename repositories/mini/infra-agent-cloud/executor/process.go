package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

const outputDrainTimeout = 5 * time.Second

func runEngineProcess(ctx context.Context, engine Engine, command EngineCommand, publish func(EventType, map[string]any) error) (*Result, error) {
	cmd := exec.CommandContext(ctx, command.Executable, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = append(cmd.Environ(), command.Env...)
	// 限制子进程退出后仍有后代持有输出管道时的等待时间。
	cmd.WaitDelay = outputDrainTimeout

	started := make(chan struct{})
	stream := func(kind EventType) *engineOutput {
		return &engineOutput{emit: func(data string) error {
			<-started
			if err := context.Cause(ctx); err != nil {
				return err
			}
			if kind == EventTypeStdout {
				engine.ParseOutput(data)
			}
			return publish(kind, map[string]any{"data": data})
		}}
	}
	stdout, stderr := stream(EventTypeStdout), stream(EventTypeStderr)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return nil, cause
		}
		return nil, fmt.Errorf("start engine process: %w", err)
	}
	// 先发送开始事件，再允许两路输出发布。
	_ = publish(EventTypeStarted, nil)
	close(started)

	// os/exec 并发读取两路管道，Wait 返回前会等待输出复制结束。
	waitErr := cmd.Wait()
	result := &Result{}
	if state := cmd.ProcessState; state != nil {
		if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			result.ExitSignal = status.Signal().String()
		} else {
			code := state.ExitCode()
			result.ExitCode = &code
		}
	}

	if err := context.Cause(ctx); err != nil {
		// 保留停止前已经收到的会话与结果信息。
		_ = engine.ParseResult(result)
		return result, err
	}
	// 进程最后一行可能没有换行符。
	if err := stdout.flush(); err != nil {
		return result, err
	}
	if err := stderr.flush(); err != nil {
		return result, err
	}
	parseErr := engine.ParseResult(result)
	if waitErr != nil {
		return result, errors.Join(fmt.Errorf("wait for engine process: %w", waitErr), parseErr)
	}
	return result, parseErr
}

// engineOutput 按行转发输出，保留跨多次管道读取的半行及换行符。
type engineOutput struct {
	pending []byte
	emit    func(string) error
}

func (w *engineOutput) Write(p []byte) (int, error) {
	size := len(p)
	for len(p) != 0 {
		end := bytes.IndexByte(p, '\n')
		if end < 0 {
			w.pending = append(w.pending, p...)
			break
		}
		w.pending = append(w.pending, p[:end+1]...)
		if err := w.flush(); err != nil {
			return size - len(p), err
		}
		p = p[end+1:]
	}
	return size, nil
}

func (w *engineOutput) flush() error {
	if len(w.pending) == 0 {
		return nil
	}
	err := w.emit(string(w.pending))
	w.pending = w.pending[:0]
	return err
}
