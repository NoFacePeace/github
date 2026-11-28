package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"infra-agent-cloud/executor"
)

func (r *Runner) startExecution(ctx context.Context, job *Job) (func() error, error) {
	spec := &executor.Spec{
		SchemaVersion:   "1",
		ExecutionID:     job.MessageID,
		TaskID:          job.TaskID,
		SessionID:       job.SessionID,
		Engine:          job.Engine,
		Prompt:          job.Prompt,
		ResumeSessionID: job.ExecutorSessionID,
		Config:          executor.Config{Model: job.Model, SystemPrompt: job.SystemPrompt},
	}
	secrets, err := r.client.ClaimSecrets(ctx, job)
	if err != nil && ctx.Err() == nil {
		return nil, r.publishJobFailure(ctx, job, fmt.Errorf("claim execution secrets: %w", err))
	}

	event := Event{Type: EventTypeStopped}
	if ctx.Err() == nil {
		handle, err := r.runtime.Start(ctx, spec, secrets)
		switch {
		case err == nil:
			return func() error {
				event, err := r.consumeExecution(ctx, job, spec, handle)
				if err != nil {
					return errors.Join(err, r.publishTerminal(ctx, job, Event{Type: EventTypeFailed, Error: err.Error()}))
				}
				return r.publishTerminal(ctx, job, event)
			}, nil
		case ctx.Err() == nil:
			event = Event{Type: EventTypeFailed, Error: err.Error()}
		}
	}
	return nil, r.publishTerminal(ctx, job, event)
}

func (r *Runner) publishJobFailure(ctx context.Context, job *Job, cause error) error {
	return r.publishTerminal(ctx, job, Event{Type: EventTypeFailed, Error: cause.Error()})
}

func (r *Runner) publishTerminal(ctx context.Context, job *Job, event Event) error {
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return r.store.PublishEvent(finalCtx, job, event)
}

type executionOutput struct {
	terminal   Event
	err        error
	publishErr error
}

func (r *Runner) consumeExecution(ctx context.Context, job *Job, spec *executor.Spec, handle *ExecutionHandle) (Event, error) {
	waitCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stdout := make(chan executionOutput, 1)
	go func() {
		output := r.readExecutionEvents(waitCtx, job, spec, handle.Stdout)
		stdout <- output
		if output.err != nil || output.publishErr != nil {
			cancel()
		}
	}()
	stderr := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.Discard, handle.Stderr)
		stderr <- err
		if err != nil {
			cancel()
		}
	}()

	exitCode, waitErr := handle.Wait(waitCtx)
	if waitErr != nil {
		stopCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		stopErr := r.runtime.Stop(stopCtx, handle)
		exitCode, waitErr = handle.WaitForExit(stopCtx)
		stop()
		if waitErr != nil {
			closeExecutionOutput(handle)
			return Event{}, fmt.Errorf("stop executor: %w", errors.Join(stopErr, waitErr))
		}
		if stopErr != nil {
			slog.Warn("Executor exited after a stop error", "executionID", job.MessageID, "error", stopErr)
		}
	}

	cleanupCtx, cleanup := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cleanup()
	defer closeExecutionOutput(handle)
	var output executionOutput
	var stderrErr error
	select {
	case output = <-stdout:
	case <-cleanupCtx.Done():
		output.err = errors.New("executor stdout did not close")
	}
	select {
	case stderrErr = <-stderr:
	case <-cleanupCtx.Done():
		stderrErr = errors.New("executor stderr did not close")
	}
	if err := r.runtime.Destroy(cleanupCtx, handle); err != nil {
		return Event{}, fmt.Errorf("destroy execution: %w", errors.Join(err, output.publishErr))
	}
	if stderrErr != nil {
		slog.Warn("Failed to drain executor stderr", "executionID", job.MessageID, "error", stderrErr)
	}
	if output.publishErr != nil {
		return Event{}, output.publishErr
	}
	if output.err == nil && output.terminal.Type == EventTypeCompleted && exitCode == 0 {
		return output.terminal, nil
	}
	if ctx.Err() != nil {
		if output.terminal.Type == EventTypeStopped {
			return output.terminal, nil
		}
		return Event{Type: EventTypeStopped}, nil
	}
	if output.err != nil {
		return Event{Type: EventTypeFailed, Error: output.err.Error()}, nil
	}
	if output.terminal.Type == EventTypeCompleted {
		return Event{Type: EventTypeFailed, Error: fmt.Sprintf("executor exited with code %d after reporting completion", exitCode)}, nil
	}
	return output.terminal, nil
}

func (r *Runner) readExecutionEvents(ctx context.Context, job *Job, spec *executor.Spec, stream io.Reader) (output executionOutput) {
	reader := bufio.NewReader(stream)
	for seq := int64(1); ; seq++ {
		line, err := reader.ReadBytes('\n')
		if len(line) == 0 && errors.Is(err, io.EOF) {
			if output.terminal.Type == "" {
				output.err = errors.New("executor exited without a terminal event")
			}
			return output
		}
		if err != nil && !errors.Is(err, io.EOF) {
			output.err = fmt.Errorf("read executor event: %w", err)
			return output
		}
		var event executor.Event
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if decoder.Decode(&event) != nil || decoder.Decode(new(json.RawMessage)) != io.EOF ||
			event.ExecutionID != spec.ExecutionID || event.SchemaVersion != spec.SchemaVersion ||
			event.Seq != seq || event.Type == "" || output.terminal.Type != "" {
			output.err = fmt.Errorf("invalid executor event at sequence %d", seq)
			return output
		}
		switch event.Type {
		case executor.EventTypeCompleted, executor.EventTypeFailed, executor.EventTypeStopped:
			result, err := json.Marshal(event.Payload["result"])
			final := Event{Type: EventType(event.Type), Output: &event}
			if err != nil || json.Unmarshal(result, &final.Result) != nil || final.Result == nil {
				output.err = errors.New("invalid executor terminal result")
				return output
			}
			final.Error = final.Result.Error
			output.terminal = final
		default:
			kind := EventTypeOutput
			if event.Type == executor.EventTypeStarted {
				kind = EventTypeStarted
			}
			if ctx.Err() == nil {
				if err := r.store.PublishEvent(ctx, job, Event{Type: kind, Output: &event}); err != nil && ctx.Err() == nil {
					output.publishErr = fmt.Errorf("publish executor event: %w", err)
					return output
				}
			}
		}
	}
}

func closeExecutionOutput(handle *ExecutionHandle) {
	for _, stream := range []io.Reader{handle.Stdout, handle.Stderr} {
		if closer, ok := stream.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}
