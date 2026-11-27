package runner

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
)

// Runner 负责接收任务和控制命令，管理执行生命周期并回传事件。
type Runner struct {
	heartbeat Heartbeat
	client    Client
	store     Store
	runtime   ProcessRuntime
	activeMu  sync.Mutex
	active    map[string]*activeExecution
}

// activeExecution 记录本轮执行的归属和取消入口，关闭 done 表示已完成收尾。
type activeExecution struct {
	job    *Job
	cancel context.CancelFunc
	done   chan struct{}
}

func NewRunner(heartbeat Heartbeat, client Client, store Store, runtime ProcessRuntime) *Runner {
	heartbeat.Engines = append([]string(nil), heartbeat.Engines...)
	heartbeat.Capacity = max(1, heartbeat.Capacity)
	return &Runner{heartbeat: heartbeat, client: client, store: store, runtime: runtime, active: make(map[string]*activeExecution)}
}

// Run 启动 Runner，使用 ctx 控制服务生命周期。
func (r *Runner) Run(ctx context.Context) error {
	if r.store == nil || r.client == nil || r.runtime == nil {
		return errors.New("runner requires a store, client and execution runtime")
	}
	if err := r.store.PublishHeartbeat(ctx, r.heartbeat); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	jobsDone := make(chan struct{})
	go func() {
		defer close(jobsDone)
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("Job consumer panicked", "panic", recovered, "stack", string(debug.Stack()))
			}
		}()
		r.consumeJobs(runCtx)
	}()
	commandsDone := make(chan struct{})
	go func() {
		defer close(commandsDone)
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("Command consumer panicked", "panic", recovered, "stack", string(debug.Stack()))
			}
		}()
		r.consumeCommands(runCtx)
	}()
	defer func() {
		cancel()
		<-jobsDone
		shutdownCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
		defer stop()
		if err := r.stopExecutions(shutdownCtx); err != nil {
			slog.Error("Failed to stop active executions", "error", err)
		}
		<-commandsDone
	}()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-jobsDone:
			if ctx.Err() != nil {
				return nil
			}
			return errors.New("job consumer exited unexpectedly")
		case <-commandsDone:
			if ctx.Err() != nil {
				return nil
			}
			return errors.New("command consumer exited unexpectedly")
		case <-ticker.C:
			if err := r.store.PublishHeartbeat(ctx, r.heartbeat); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				slog.Error("Failed to publish runner heartbeat", "runnerID", r.heartbeat.RunnerID, "error", err)
			}
		}
	}
}

func (r *Runner) consumeJobs(ctx context.Context) {
	for ctx.Err() == nil {
		job, entryID, err := r.store.ClaimJob(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Error("Failed to claim runner job", "error", err)
		} else if job != nil {
			handleErr := r.handleJob(ctx, job)
			ackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			if handleErr != nil {
				slog.Error("Failed to handle runner job", "error", handleErr)
				if err := r.store.NackJob(ackCtx, entryID, handleErr.Error()); err != nil {
					slog.Error("Failed to retain runner job for recovery", "error", err)
				}
			} else {
				err := r.store.AckJob(ackCtx, entryID)
				cancel()
				if err == nil {
					continue
				}
				slog.Error("Failed to acknowledge runner job", "error", err)
			}
			cancel()
		}
		// 暂停轮询，避免空队列或持续错误导致忙循环。
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (r *Runner) consumeCommands(ctx context.Context) {
	for ctx.Err() == nil {
		command, err := r.store.ClaimCommand(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Error("Failed to claim runner command", "error", err)
		} else if command != nil {
			if err := r.handleCommand(ctx, command); err != nil {
				slog.Error("Failed to handle runner command", "error", err)
				if nackErr := r.store.NackCommand(ctx, command, err.Error()); nackErr != nil {
					slog.Error("Failed to retain runner command for recovery", "error", nackErr)
				}
			} else if err := r.store.AckCommand(ctx, command); err != nil {
				slog.Error("Failed to acknowledge runner command", "error", err)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (r *Runner) handleCommand(ctx context.Context, command *Command) error {
	if command.RunnerID != r.heartbeat.RunnerID || !time.Now().Before(command.ExpiresAt) {
		return nil
	}
	if command.Type == CommandTypeStop {
		return r.stopExecution(ctx, command)
	}
	return errors.New("unsupported runner command")
}

func (r *Runner) handleJob(ctx context.Context, job *Job) error {
	if job.MessageID == "" || job.TaskID == "" || job.SessionID == "" ||
		job.RunnerID == "" || job.RunnerID != r.heartbeat.RunnerID || job.ExecutionToken == "" {
		return errors.New("invalid runner job identity")
	}
	executionCtx, cancel := context.WithCancel(ctx)
	unregister, err := r.registerExecution(job, cancel)
	if err != nil || unregister == nil {
		cancel()
		return err
	}
	background := false
	var stopRenewal func()
	finish := func() {
		cancel()
		if stopRenewal != nil {
			stopRenewal()
		}
		unregister()
	}
	defer func() {
		if !background {
			finish()
		}
	}()
	stopRenewal, err = r.startExecutionRenewal(executionCtx, job, cancel)
	if err != nil {
		return err
	}
	complete, err := r.startExecution(executionCtx, job)
	if err != nil || complete == nil {
		return err
	}

	// 启动成功后把收尾交给后台，消费循环可以确认消息并领取下一条。
	go func() {
		defer finish()
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("Execution panicked", "executionID", job.MessageID, "panic", recovered, "stack", string(debug.Stack()))
			}
		}()
		if err := complete(); err != nil {
			slog.Error("Failed to finish execution", "executionID", job.MessageID, "error", err)
		}
	}()
	background = true
	return nil
}

func (r *Runner) registerExecution(job *Job, cancel context.CancelFunc) (func(), error) {
	r.activeMu.Lock()
	defer r.activeMu.Unlock()
	if active := r.active[job.SessionID]; active != nil &&
		active.job.MessageID == job.MessageID && active.job.TaskID == job.TaskID &&
		active.job.ExecutionToken == job.ExecutionToken {
		// 重投沿用已经启动的执行，消费循环可以再次确认同一条消息。
		return nil, nil
	}
	for _, active := range r.active {
		if active.job.TaskID == job.TaskID || active.job.SessionID == job.SessionID {
			return nil, errors.New("task or session is already executing")
		}
	}
	if len(r.active) >= r.heartbeat.Capacity {
		return nil, errors.New("runner capacity exhausted")
	}
	active := &activeExecution{job: job, cancel: cancel, done: make(chan struct{})}
	r.active[job.SessionID] = active
	return func() {
		r.activeMu.Lock()
		defer r.activeMu.Unlock()
		delete(r.active, job.SessionID)
		close(active.done)
	}, nil
}

func (r *Runner) stopExecution(ctx context.Context, command *Command) error {
	r.activeMu.Lock()
	active := r.active[command.SessionID]
	r.activeMu.Unlock()
	if active == nil || active.job.TaskID != command.TaskID || active.job.ExecutionToken != command.ExecutionToken {
		return nil
	}
	active.cancel()
	select {
	case <-active.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) stopExecutions(ctx context.Context) error {
	r.activeMu.Lock()
	active := make([]*activeExecution, 0, len(r.active))
	for _, execution := range r.active {
		execution.cancel()
		active = append(active, execution)
	}
	r.activeMu.Unlock()
	for _, execution := range active {
		select {
		case <-execution.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
