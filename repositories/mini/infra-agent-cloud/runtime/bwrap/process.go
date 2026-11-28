package bwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"infra-agent-cloud/executor"
	"infra-agent-cloud/runtime"
)

func startBwrapProcess(ctx context.Context, cmd *exec.Cmd, spec *executor.Spec, secrets *executor.Secrets, cgroup, childPIDFile string) (_ *runtime.Execution, _ <-chan struct{}, err error) {
	input := executor.Secrets{}
	if secrets != nil {
		input = *secrets
	}
	if input.ExecutionID == "" {
		input.ExecutionID = spec.ExecutionID
	}
	if input.ExecutionID != spec.ExecutionID {
		return nil, nil, errors.New("executor secrets have the wrong execution ID")
	}
	frame, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("encode executor secrets: %w", err)
	}

	stdout, stdoutWriter, err := openBwrapOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("open executor stdout: %w", err)
	}
	defer stdoutWriter.Close()
	stderr, stderrWriter, err := openBwrapOutput()
	if err != nil {
		_ = stdout.Close()
		return nil, nil, fmt.Errorf("open executor stderr: %w", err)
	}
	defer stderrWriter.Close()
	defer func() {
		if err != nil {
			_ = stdout.Close()
			_ = stderr.Close()
		}
	}()
	cmd.Stdout, cmd.Stderr = stdoutWriter, stderrWriter
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("open executor stdin: %w", err)
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, nil, fmt.Errorf("start executor process: %w", err)
	}
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	var exitCode int
	var waitErr error
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		waitErr = cmd.Wait()
		exitCode = -1
		if state := cmd.ProcessState; state != nil {
			exitCode = state.ExitCode()
			if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				exitCode = 128 + int(status.Signal())
			}
		}
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			waitErr = nil
		}
	}()

	inputSlot := make(chan struct{}, 1)
	sendInput := func(ctx context.Context, input string) error {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case inputSlot <- struct{}{}:
			defer func() { <-inputSlot }()
		}
		if err := context.Cause(ctx); err != nil {
			return err
		}
		// 中断写入时关闭管道，避免将下一帧接在未写完的帧后面。
		stop := context.AfterFunc(ctx, func() { _ = stdin.Close() })
		defer stop()
		_, err := io.WriteString(stdin, input)
		if cause := context.Cause(ctx); cause != nil {
			return cause
		}
		return err
	}
	inputErr := sendInput(ctx, string(frame)+"\n")
	steer := spec.SteerPromptMode == executor.SteerPromptModeSteer
	if inputErr == nil && !steer {
		if err := stdin.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			inputErr = err
		}
	}
	var interruptPID int
	if inputErr != nil {
		err = fmt.Errorf("write executor secrets: %w", inputErr)
	} else {
		interruptPID, err = readBwrapChildPID(ctx, childPIDFile, exited)
	}
	if err != nil {
		_ = stdin.Close()
		killErr := cmd.Process.Kill()
		if errors.Is(killErr, os.ErrProcessDone) {
			killErr = nil
		}
		if cgroup != "" {
			killErr = errors.Join(killErr, killBwrapCgroup(cgroup))
		}
		<-exited
		return nil, nil, errors.Join(err, killErr)
	}

	waitForExit := func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return -1, context.Cause(ctx)
		case <-exited:
			return exitCode, waitErr
		}
	}
	handle := &runtime.Execution{
		ExecutionID:  spec.ExecutionID,
		PID:          cmd.Process.Pid,
		InterruptPID: interruptPID,
		Stdout:       stdout,
		Stderr:       stderr,
		WaitForExit:  waitForExit,
		Wait: func(ctx context.Context) (int, error) {
			code, err := waitForExit(ctx)
			if err != nil {
				return code, err
			}
			for _, done := range []<-chan struct{}{stdout.done, stderr.done} {
				select {
				case <-ctx.Done():
					return -1, context.Cause(ctx)
				case <-done:
				}
			}
			return code, nil
		},
	}
	if steer {
		handle.SendInput = sendInput
	}
	return handle, exited, nil
}

type bwrapOutput struct {
	file *os.File
	done chan struct{}
	once sync.Once
}

func openBwrapOutput() (*bwrapOutput, *os.File, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return &bwrapOutput{file: reader, done: make(chan struct{})}, writer, nil
}

func (o *bwrapOutput) Read(p []byte) (int, error) {
	select {
	case <-o.done:
		return 0, io.EOF
	default:
	}
	n, err := o.file.Read(p)
	if err != nil {
		_ = o.Close()
	}
	return n, err
}

func (o *bwrapOutput) Close() (err error) {
	o.once.Do(func() {
		err = o.file.Close()
		close(o.done)
	})
	return err
}

func killBwrapCgroup(dir string) error {
	file, err := os.OpenFile(filepath.Join(dir, "cgroup.kill"), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open cgroup.kill: %w", err)
	}
	_, writeErr := io.WriteString(file, "1")
	return errors.Join(writeErr, file.Close())
}
