package runner

import (
	"context"
	"errors"
	"infra-agent-cloud/executor"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
)

// Runner 负责接收任务和控制命令，管理执行生命周期并回传事件。
type Runner struct {
	heartbeat Heartbeat
	client    Client
	runtime   Runtime
	activeMu  sync.Mutex
	active    map[string]*activeExecution
}

// activeExecution 管理节点内的执行句柄，Session 标识和执行 token 的索引规则后续补充。
type activeExecution struct {
	job    *Job
	cancel context.CancelFunc
	done   chan struct{}
}

func NewRunner(heartbeat Heartbeat, client Client, runtime Runtime) *Runner {
	heartbeat.Engines = append([]string(nil), heartbeat.Engines...)
	return &Runner{heartbeat: heartbeat, client: client, runtime: runtime, active: make(map[string]*activeExecution)}
}

// Run 启动 Runner，使用 ctx 控制服务生命周期。
func (r *Runner) Run(ctx context.Context) error {
	if err := r.client.PublishHeartbeat(ctx, r.heartbeat); err != nil {
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
		shutdownCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer stop()
		if err := r.stopExecutions(shutdownCtx); err != nil {
			slog.Error("Failed to stop active executions", "error", err)
		}
		<-jobsDone
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
			if err := r.client.PublishHeartbeat(ctx, r.heartbeat); err != nil {
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
		job, err := r.client.ClaimJob(ctx)
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
				if err := r.client.NackJob(ackCtx, job, handleErr.Error()); err != nil {
					slog.Error("Failed to retain runner job for recovery", "error", err)
				}
			} else if err := r.client.AckJob(ackCtx, job); err != nil {
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
		command, err := r.client.ClaimCommand(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Error("Failed to claim runner command", "error", err)
		} else if command != nil {
			if err := r.handleCommand(ctx, command); err != nil {
				slog.Error("Failed to handle runner command", "error", err)
				if nackErr := r.client.NackCommand(ctx, command, err.Error()); nackErr != nil {
					slog.Error("Failed to retain runner command for recovery", "error", nackErr)
				}
			} else if err := r.client.AckCommand(ctx, command); err != nil {
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
	if command.Type == CommandTypeStop {
		return r.stopExecution(ctx, command)
	}
	return errors.New("unsupported runner command")
}

func (r *Runner) handleJob(ctx context.Context, job *Job) error {
	// TODO: 校验消息协议、目标 Runner、执行 token 和重复执行。
	executionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	unregister, err := r.registerExecution(job, cancel)
	if err != nil {
		return err
	}
	defer unregister()
	stopRenewal, err := r.startExecutionRenewal(executionCtx, job, cancel)
	if err != nil {
		return err
	}
	defer stopRenewal()
	return r.runExecution(executionCtx, job)
}

func (r *Runner) runExecution(ctx context.Context, job *Job) error {
	// started 当前表示进入执行准备，进程就绪事件的准确边界后续补充。
	if err := r.client.PublishEvent(ctx, job, Event{Type: EventTypeStarted}); err != nil {
		return err
	}
	result, err := r.runtime.Run(ctx, job, func(output executor.Event) error {
		return r.client.PublishEvent(ctx, job, Event{Type: EventTypeOutput, Output: &output})
	})
	event := Event{Type: EventTypeCompleted, Result: result}
	if err != nil {
		event.Type = EventTypeFailed
		event.Error = err.Error()
	}
	if ctx.Err() != nil {
		event.Type = EventTypeStopped
	}
	// 取消后仍尝试回传终态，成功回传后才允许确认消费。
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return r.client.PublishEvent(finalCtx, job, event)
}

func (r *Runner) registerExecution(job *Job, cancel context.CancelFunc) (func(), error) {
	// TODO: 加锁登记活跃执行，校验执行 token、去重并限制本地并发；返回注销及完成通知函数。
	return nil, errors.New("execution registration is not implemented")
}

func (r *Runner) startExecutionRenewal(ctx context.Context, job *Job, cancel context.CancelFunc) (func(), error) {
	// TODO: 通过 Client.RenewExecution 首次校验并持续续约，明确拒绝或到期时取消执行。
	// 临时通信失败允许在租约期限内重试，返回的函数须停止并等待续约循环退出。
	return nil, errors.New("execution renewal is not implemented")
}

func (r *Runner) stopExecution(ctx context.Context, command *Command) error {
	// TODO: 按目标执行 token 查找活跃执行，取消并等待退出；尚在准备的执行也须记录停止意图。
	return errors.New("execution stop is not implemented")
}

func (r *Runner) stopExecutions(ctx context.Context) error {
	// TODO: 停止接单后取消并等待全部活跃执行，超过退出期限时报告未退出的进程。
	return errors.New("execution shutdown is not implemented")
}
