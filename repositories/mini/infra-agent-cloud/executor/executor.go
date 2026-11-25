package executor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
)

// Run 是公共执行入口，commands 为可选的执行中输入，普通模式传 nil。
func Run(ctx context.Context, spec *Spec, secrets *Secrets, commands <-chan SteerCommand, emit func(Event) error) (*Result, error) {
	publisher := eventPublisher{
		schemaVersion: spec.SchemaVersion,
		executionID:   spec.ExecutionID,
		writeEvent:    emit,
	}
	result, runErr := runEngine(ctx, spec, secrets, publisher.publish)
	if result == nil {
		result = &Result{}
	}
	kind := EventTypeCompleted
	if errors.Is(ctx.Err(), context.Canceled) {
		kind = EventTypeStopped
		runErr = context.Cause(ctx)
	} else if runErr != nil {
		kind = EventTypeFailed
		result.Error = runErr.Error()
	}
	// 取消后仍尝试写出终态。若先前写入已失败，发布器会返回首次错误。
	if err := publisher.publish(kind, map[string]any{"result": result}); err != nil {
		return result, err
	}
	return result, runErr
}

func runEngine(ctx context.Context, spec *Spec, secrets *Secrets, publish func(EventType, map[string]any) error) (*Result, error) {
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	var engine Engine
	switch spec.Engine {
	case "codex":
		engine = NewCodex()
	default:
		return nil, fmt.Errorf("unsupported engine: %q", spec.Engine)
	}
	if err := publish(EventTypeStarted, nil); err != nil {
		return nil, err
	}

	command, err := engine.BuildCommand(spec, secrets)
	if err != nil {
		return nil, fmt.Errorf("build %s command: %w", spec.Engine, err)
	}
	path, err := exec.LookPath(command.Executable)
	if err != nil {
		return nil, fmt.Errorf("cannot find %s executable: %w", command.Executable, err)
	}
	command.Executable = path
	if err := engine.WriteConfig(spec, secrets); err != nil {
		return nil, fmt.Errorf("prepare %s config: %w", spec.Engine, err)
	}

	return runEngineProcess(ctx, engine, command, spec, publish)
}

func runEngineProcess(ctx context.Context, engine Engine, command EngineCommand, spec *Spec, publish func(EventType, map[string]any) error) (result *Result, runErr error) {
	sessionLog, err := openSessionLog(spec)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := sessionLog.close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close session log: %w", err))
		}
	}()

	cmd := exec.CommandContext(ctx, command.Executable, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = append(cmd.Environ(), command.Env...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create engine stdout pipe: %w", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		if writer, ok := cmd.Stdout.(io.Closer); ok {
			_ = writer.Close()
		}
		return nil, fmt.Errorf("create engine stderr pipe: %w", err)
	}
	defer stderr.Close()

	var stdoutLog strings.Builder
	var sessionID string
	var mcpStatus *EngineMCPStatus
	outputErrors := make(chan error, 2)
	stream := func(pipe io.ReadCloser, kind EventType, filter func(string) *string) {
		err := readEngineOutput(pipe, func(data string) error {
			if err := context.Cause(ctx); err != nil {
				return err
			}
			if kind == EventTypeStdout {
				stdoutLog.WriteString(data)
				if id := engine.ParseSessionID(data); id != "" {
					sessionID = id
				}
			}
			logLine := &data
			if filter != nil {
				logLine = filter(data)
			}
			if err := sessionLog.append(logLine); err != nil {
				return err
			}
			payload := map[string]any{"data": data}
			if logLine == nil {
				payload["logData"] = ""
			} else if *logLine != data {
				payload["logData"] = *logLine
			}
			if err := publish(kind, payload); err != nil {
				return err
			}
			if kind != EventTypeStdout {
				return nil
			}
			if status := engine.ParseMCPStatus(data); status != nil {
				mcpStatus = status
			}
			for _, output := range engine.ParseChannelOutput(data) {
				payload := map[string]any{"kind": output.Kind, "text": output.Text}
				if output.Replace {
					payload["replace"] = true
				}
				if err := publish(EventTypeChannelOutput, payload); err != nil {
					return err
				}
			}
			return nil
		})
		_ = pipe.Close()
		if err != nil {
			err = fmt.Errorf("read engine %s: %w", kind, err)
		}
		outputErrors <- err
	}
	if err := cmd.Start(); err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return nil, cause
		}
		return nil, fmt.Errorf("start engine process: %w", err)
	}
	var stdoutFilter, stderrFilter func(string) *string
	if filter, ok := engine.(EngineLogFilter); ok {
		stdoutFilter = filter.FilterLogLine
	}
	if filter, ok := engine.(EngineErrorLogFilter); ok {
		stderrFilter = filter.FilterErrorLine
	}
	go stream(stdout, EventTypeStdout, stdoutFilter)
	go stream(stderr, EventTypeStderr, stderrFilter)
	// Wait 会关闭输出管道，须等两路读取结束后再调用。
	outputErr := errors.Join(<-outputErrors, <-outputErrors)
	waitErr := cmd.Wait()

	result = &Result{}
	if state := cmd.ProcessState; state != nil {
		if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			result.ExitSignal = status.Signal().String()
		} else {
			code := state.ExitCode()
			result.ExitCode = &code
		}
	}

	cause := context.Cause(ctx)
	parsed, parseErr := engine.ParseResult(stdoutLog.String(), spec)
	result.ExecutorSessionID = sessionID
	result.MCP = mcpStatus
	result.Text = parsed.Text
	result.Usage = parsed.Usage
	if cause != nil {
		return result, cause
	}
	if waitErr != nil {
		waitErr = fmt.Errorf("wait for engine process: %w", waitErr)
	}
	return result, errors.Join(waitErr, outputErr, parseErr)
}

// readEngineOutput 逐行转发输出，保留换行符及最后一行的内容。
func readEngineOutput(input io.Reader, emit func(string) error) error {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if emitErr := emit(line); emitErr != nil {
				return emitErr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
