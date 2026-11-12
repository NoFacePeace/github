package server

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"
	"time"
)

// SaveDraftInput 保存草稿的输入，配置字段后续补充。
type SaveDraftInput struct{}

// InvokeTaskInput 定义任务执行输入，TaskID 为空时创建任务，否则继续已有任务。
type InvokeTaskInput struct {
	AgentID string
	TaskID  string
	Prompt  string
}

// Service 定义 Agent 管理、草稿编辑、版本发布和任务执行的业务接口。
type Service interface {
	// CreateAgent 创建 Agent，同时创建初始草稿。
	CreateAgent(ctx context.Context) (*Agent, error)
	GetAgent(ctx context.Context, agentID string) (*Agent, error)
	ListAgents(ctx context.Context) ([]Agent, error)

	GetDraft(ctx context.Context, agentID string) (*AgentVersion, error)
	// SaveDraft 保存草稿，有则更新，无则创建。
	SaveDraft(ctx context.Context, agentID string, input SaveDraftInput) error

	// SaveVersion 将当前草稿保存为不可修改的版本快照，保留草稿。
	SaveVersion(ctx context.Context, agentID string) (*AgentVersion, error)
	GetVersion(ctx context.Context, versionID string) (*AgentVersion, error)
	ListVersions(ctx context.Context, agentID string) ([]AgentVersion, error)

	// PublishVersion 校验版本归属，并更新 Agent 的当前发布版本。
	PublishVersion(ctx context.Context, agentID, versionID string) error

	// InvokeTask 创建任务或继续已有任务，并发起一轮执行。
	InvokeTask(ctx context.Context, input InvokeTaskInput) (*Task, error)
}

type service struct {
	store         Store
	schedulerWake chan struct{}
}

// NewService 创建服务并启动调度循环，ctx 应使用服务生命周期的上下文。
func NewService(ctx context.Context, store Store) *service {
	s := &service{
		store:         store,
		schedulerWake: make(chan struct{}, 1),
	}
	go s.runScheduler(ctx)
	return s
}

func (s *service) InvokeTask(ctx context.Context, input InvokeTaskInput) (*Task, error) {
	if input.TaskID != "" {
		return s.resumeTask(ctx, input)
	}
	return s.createTask(ctx, input)
}

func (s *service) createTask(ctx context.Context, input InvokeTaskInput) (*Task, error) {
	agent, err := s.store.GetAgent(ctx, input.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.PublishedVersionID == "" {
		return nil, errors.New("agent has no published version")
	}
	version, err := s.store.GetVersion(ctx, agent.PublishedVersionID)
	if err != nil {
		return nil, err
	}
	if version.AgentID != agent.ID || version.Status != AgentVersionStatusSaved {
		return nil, errors.New("invalid published agent version")
	}

	task := &Task{
		ID:             rand.Text(),
		AgentID:        agent.ID,
		AgentVersionID: version.ID,
	}
	if err := s.store.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	session := &Session{
		ID:        rand.Text(),
		TaskID:    task.ID,
		Prompt:    input.Prompt,
		Status:    SessionStatusRunning,
		CreatedAt: time.Now(),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	if err := s.runSession(ctx, task, session, version); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *service) resumeTask(ctx context.Context, input InvokeTaskInput) (*Task, error) {
	task, err := s.store.GetTask(ctx, input.TaskID)
	if err != nil {
		return nil, err
	}
	lastSession, err := s.store.GetLastSession(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	if lastSession != nil && lastSession.Status == SessionStatusRunning {
		return nil, errors.New("task already has an active session")
	}
	// 继续任务始终沿用 Task 绑定的版本，不读取 Agent 当前发布版本或草稿。
	version, err := s.store.GetVersion(ctx, task.AgentVersionID)
	if err != nil {
		return nil, err
	}
	session := &Session{
		ID:        rand.Text(),
		TaskID:    task.ID,
		Prompt:    input.Prompt,
		Status:    SessionStatusRunning,
		CreatedAt: time.Now(),
	}
	if lastSession != nil {
		session.RunnerID = lastSession.RunnerID
		session.ExecutorSessionID = lastSession.ExecutorSessionID
	}
	// CreateSession 的存储实现须通过活跃会话唯一约束保证并发安全。
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	if err := s.runSession(ctx, task, session, version); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *service) runSession(ctx context.Context, task *Task, session *Session, version *AgentVersion) error {
	job := &Job{
		TaskID:            task.ID,
		SessionID:         session.ID,
		RunnerID:          session.RunnerID,
		ExecutorSessionID: session.ExecutorSessionID,
		Prompt:            session.Prompt,
		AgentVersionID:    task.AgentVersionID,
		Engine:            version.Engine,
		Model:             version.Model,
		SystemPrompt:      version.SystemPrompt,
	}
	if err := s.submitJob(ctx, job); err != nil {
		// 请求取消后仍尝试保存失败状态，但限制收尾耗时。
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		failErr := s.failSession(cleanupCtx, session.ID, err.Error())
		if failErr != nil {
			return errors.Join(err, fmt.Errorf("fail session after job submission error: %w", failErr))
		}
		return err
	}
	s.wakeScheduler()
	return nil
}

func (s *service) failSession(ctx context.Context, sessionID, reason string) error {
	if err := s.store.FailSession(ctx, sessionID, reason); err != nil {
		return err
	}
	job, err := s.store.GetJob(ctx, sessionID)
	if err != nil {
		return err
	}
	if job == nil || job.RunnerID == "" || job.ExecutionToken == "" {
		return nil
	}
	if !time.Now().Before(job.ExecutionExpiresAt) {
		return nil
	}
	return s.store.PublishCommand(ctx, &RunnerCommand{
		ID:             rand.Text(),
		Type:           RunnerCommandTypeStop,
		RunnerID:       job.RunnerID,
		TaskID:         job.TaskID,
		SessionID:      job.SessionID,
		ExecutionToken: job.ExecutionToken,
		ExpiresAt:      job.ExecutionExpiresAt,
	})
}

func (s *service) submitJob(ctx context.Context, job *Job) error {
	if job.MessageID == "" {
		job.MessageID = rand.Text()
	}
	if job.Status == "" {
		job.Status = JobStatusQueued
	}
	if job.QueuedAt.IsZero() {
		job.QueuedAt = time.Now()
	}
	if job.Deadline.IsZero() {
		job.Deadline = job.QueuedAt.Add(10 * time.Minute)
	}
	return s.store.EnqueueJob(ctx, job)
}

func (s *service) wakeScheduler() {
	select {
	case s.schedulerWake <- struct{}{}:
	default:
		// 已有待处理的唤醒通知，合并本次通知。
	}
}

func (s *service) runScheduler(ctx context.Context) {
	for ctx.Err() == nil {
		if s.runSchedulerOnce(ctx) {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}

		// 重启后主动扫描，避免此前的唤醒通知已经被消费。
		s.wakeScheduler()
	}
}

// runSchedulerOnce 正常退出返回 true，发生 panic 返回 false。
func (s *service) runSchedulerOnce(ctx context.Context) (stopped bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("Scheduler panicked", "panic", recovered, "stack", string(debug.Stack()))
		}
	}()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return true
		case <-s.schedulerWake:
		case <-ticker.C:
			if err := s.recoverSessions(ctx); err != nil {
				slog.Error("Failed to recover sessions", "error", err)
			}
		}
		if err := s.dispatchPending(ctx); err != nil {
			slog.Error("Failed to dispatch pending jobs", "error", err)
		}
	}
}

func (s *service) recoverSessions(ctx context.Context) error {
	startupBefore := time.Now().Add(-5 * time.Minute)
	cursor := ""
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		sessions, nextCursor, err := s.store.ListRecoverableSessions(ctx, startupBefore, cursor, 100)
		if err != nil {
			return err
		}
		slog.Debug("Session recovery candidates loaded", "count", len(sessions))
		for _, session := range sessions {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := s.recoverSession(ctx, session.ID, startupBefore); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				slog.Warn("Session recovery requires retry", "sessionID", session.ID, "error", err)
			}
		}
		if nextCursor == "" {
			return nil
		}
		cursor = nextCursor
	}
}

// recoverSession 复查单条候选，调度未完成时交给调度器，已投递后的异常按执行租约处理。
// 当前跳过事件续期和终态事件回放，已有业务终态保持不变。
func (s *service) recoverSession(ctx context.Context, sessionID string, startupBefore time.Time) error {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil || session.Status == SessionStatusCompleted {
		return nil
	}
	job, err := s.store.GetJob(ctx, sessionID)
	if err != nil {
		return err
	}
	if job != nil && job.Status != JobStatusDispatched {
		// 排队、投递重试及调度终态同步均由 pending 调度流程处理。
		s.wakeScheduler()
		return nil
	}
	lease, err := s.store.GetExecutionLease(ctx, sessionID)
	if err != nil {
		return err
	}
	if lease != nil {
		if session.Status != SessionStatusFailed && session.Status != SessionStatusStopped {
			return nil
		}
		// 按当前租约补发幂等停止命令；不提前释放执行租约和容量。
		return s.store.PublishCommand(ctx, &RunnerCommand{
			ID:             "stop:" + lease.ExecutionToken,
			Type:           RunnerCommandTypeStop,
			RunnerID:       lease.RunnerID,
			TaskID:         lease.TaskID,
			SessionID:      lease.SessionID,
			ExecutionToken: lease.ExecutionToken,
			ExpiresAt:      lease.ExecutionExpiresAt,
		})
	}
	if session.Status != SessionStatusRunning {
		return nil
	}
	if session.CreatedAt.IsZero() || session.CreatedAt.After(startupBefore) {
		return nil
	}
	// 写入前由 Store 复查，避免扫描期间已有终态或新租约时被误标为失败。
	_, err = s.store.FailSessionIfLeaseExpired(ctx, sessionID, startupBefore, "execution lease expired")
	return err
}

func (s *service) dispatchPending(ctx context.Context) error {
	sessionIDs, err := s.store.ListPendingJobIDs(ctx)
	if err != nil {
		return err
	}
	for _, sessionID := range sessionIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.dispatchPendingJob(ctx, sessionID); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// 单条失败保留待调度记录，继续处理其他 Job。
			slog.Warn("Job dispatch requires recovery", "sessionID", sessionID, "error", err)
		}
	}
	return nil
}

// dispatchPendingJob 处理单条 Job，持锁期间的所有返回路径都会释放本次调度锁。
func (s *service) dispatchPendingJob(ctx context.Context, sessionID string) (err error) {
	claimToken, err := s.store.ClaimJob(ctx, sessionID, 30*time.Second)
	if err != nil {
		return err
	}
	if claimToken == "" {
		return nil
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		err = errors.Join(err, s.store.ReleaseJobClaim(releaseCtx, sessionID, claimToken))
	}()
	renewCtx, stopRenewal, err := s.startJobClaimRenewal(ctx, sessionID, claimToken)
	if err != nil {
		return err
	}
	ctx = renewCtx
	// 先停止并等待续约退出，再由前面的 defer 释放调度锁。
	defer func() {
		err = errors.Join(err, stopRenewal())
	}()
	job, err := s.store.GetJob(ctx, sessionID)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	job.ClaimToken = claimToken
	if job.Status == JobStatusQueued {
		// 准备前同步排队状态，业务层可能根据最新 Session 取消排队。
		if _, err := s.syncSessionDispatch(ctx, job); err != nil {
			return err
		}
		latest, err := s.store.GetJob(ctx, sessionID)
		if err != nil {
			return err
		}
		if latest == nil || latest.Status != JobStatusQueued {
			return nil
		}
		job = latest
		job.ClaimToken = claimToken
	}
	if job.Status == JobStatusQueued && !job.Deadline.IsZero() && !time.Now().Before(job.Deadline) {
		failed := *job
		failed.Status = JobStatusFailed
		failed.Error = "job dispatch deadline exceeded"
		// 先保存调度失败，再由后面的统一入口同步 Session。
		saved, err := s.store.SaveJob(ctx, &failed, claimToken)
		if err != nil {
			return err
		}
		if !saved {
			return nil
		}
		*job = failed
	}
	if job.Status == JobStatusQueued {
		lease, err := s.store.GetExecutionLease(ctx, job.SessionID)
		if err != nil {
			return err
		}
		if lease != nil {
			// queued 尚无可发布的固定消息，先清理遗留预占。
			released, err := s.store.ReleaseExecutionLease(ctx, job.SessionID, lease.ExecutionToken)
			if err != nil {
				return err
			}
			if !released {
				return nil
			}
		}
		runners, err := s.store.ListRunners(ctx)
		if err != nil {
			return err
		}
		var selected *Runner
		for i := range runners {
			runner := &runners[i]
			if job.RunnerID != "" && runner.ID != job.RunnerID {
				continue
			}
			if runner.Status != RunnerStatusOnline ||
				!slices.Contains(runner.Engines, job.Engine) ||
				(job.RunnerID == "" && runner.ActiveSessions >= runner.Capacity) {
				continue
			}
			selected = runner
			break
		}
		if selected == nil {
			return errors.New("no available compatible runner")
		}
		lease, err = s.store.ReserveExecutionLease(ctx, job.TaskID, job.SessionID, selected.ID, selected.Capacity, 5*time.Minute)
		if err != nil {
			return err
		}
		job.RunnerID = lease.RunnerID
		job.ExecutionToken = lease.ExecutionToken
		job.ExecutionExpiresAt = lease.ExecutionExpiresAt
		job.Status = JobStatusDispatching
		// 固定消息持久化成功后才允许投递；保存结果不确定时保留租约供后续恢复。
		saved, err := s.store.SaveJob(ctx, job, claimToken)
		if err != nil {
			return err
		}
		if !saved {
			return nil
		}
	}
	if job.Status == JobStatusDispatching {
		// 固定 Job 保存后更新 Session 绑定；重试时同样补写，成功后才投递。
		if err := s.store.UpdateSessionRunner(ctx, job.SessionID, job.RunnerID); err != nil {
			return err
		}
		dispatchErr := s.dispatchJob(ctx, job)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		result := *job
		result.Status = JobStatusDispatched
		if dispatchErr != nil {
			if !errors.Is(dispatchErr, ErrExecutionLeaseExpired) && !errors.Is(dispatchErr, ErrDispatchConflict) {
				return dispatchErr
			}
			// 确定失败先保存调度结果，再交给后面的业务同步和收尾；其他错误保留原阶段重试。
			result.Status = JobStatusFailed
			result.Error = dispatchErr.Error()
		}
		saved, err := s.store.SaveJob(ctx, &result, claimToken)
		if err != nil {
			return err
		}
		if !saved {
			return errors.New("job dispatch confirmation rejected")
		}
		*job = result
	}
	settled, err := s.syncSessionDispatch(ctx, job)
	if err != nil {
		return err
	}
	if !settled {
		return nil
	}
	if job.Status == JobStatusDispatched || job.Status == JobStatusFailed || job.Status == JobStatusCancelled {
		_, err = s.store.AcknowledgeJobDispatch(ctx, job.SessionID, claimToken)
		return err
	}
	return nil
}

// syncSessionDispatch 统一复查最新 Session、同步调度摘要，并处理失败、取消排队及停止补偿。
// 调度层须先分类异常并保存 Job 结果，本方法仅同步已形成的调度状态。
// 返回 true 表示业务处理完成，允许调度收尾；false 表示仍需补偿，保留待调度记录。
func (s *service) syncSessionDispatch(ctx context.Context, job *Job) (bool, error) {
	session, err := s.store.UpdateSessionDispatch(ctx, job.SessionID, &SessionDispatch{
		MessageID: job.MessageID,
		Status:    job.Status,
		QueuedAt:  job.QueuedAt,
		Deadline:  job.Deadline,
		Error:     job.Error,
	})
	if err != nil {
		return false, err
	}
	if session == nil {
		return false, errors.New("session not found")
	}
	if session.Status == SessionStatusRunning {
		return true, nil
	}
	// queued 尚未形成固定执行消息；没有执行 token 的失败或取消也只可能留下未发布的预占。
	unpublished := job.Status == JobStatusQueued ||
		((job.Status == JobStatusFailed || job.Status == JobStatusCancelled) && job.ExecutionToken == "")
	if job.Status == JobStatusQueued {
		cancelled := *job
		cancelled.Status = JobStatusCancelled
		saved, err := s.store.SaveJob(ctx, &cancelled, job.ClaimToken)
		if err != nil {
			return false, err
		}
		if !saved {
			return false, nil
		}
		*job = cancelled
		s.wakeScheduler()
	}
	if !unpublished && session.Status == SessionStatusCompleted {
		return true, nil
	}
	lease, err := s.store.GetExecutionLease(ctx, job.SessionID)
	if err != nil {
		return false, err
	}
	if lease == nil {
		return true, nil
	}
	if unpublished {
		// 明确尚未发布的预占可以直接释放；token 已变化时保留记录，交给后续扫描复查。
		released, err := s.store.ReleaseExecutionLease(ctx, job.SessionID, lease.ExecutionToken)
		if err != nil {
			return false, err
		}
		if released {
			s.wakeScheduler()
		}
		return released, nil
	}
	// 已进入投递阶段时按当前租约发送停止命令，重试复用命令 ID，不提前释放执行容量。
	if err := s.store.PublishCommand(ctx, &RunnerCommand{
		ID:             "stop:" + lease.ExecutionToken,
		Type:           RunnerCommandTypeStop,
		RunnerID:       lease.RunnerID,
		TaskID:         lease.TaskID,
		SessionID:      lease.SessionID,
		ExecutionToken: lease.ExecutionToken,
		ExpiresAt:      lease.ExecutionExpiresAt,
	}); err != nil {
		return false, err
	}
	return false, nil
}

// startJobClaimRenewal 在读取 Job 前确认调度锁并启动续约，失去锁时取消返回的 Context。
func (s *service) startJobClaimRenewal(ctx context.Context, sessionID, claimToken string) (context.Context, func() error, error) {
	const claimTTL = 30 * time.Second
	const renewalInterval = 10 * time.Second

	renewCtx, cancel := context.WithCancelCause(ctx)
	renew := func() error {
		requestCtx, stop := context.WithTimeout(renewCtx, renewalInterval)
		defer stop()
		ok, err := s.store.RenewJobClaim(requestCtx, sessionID, claimToken, claimTTL)
		if err != nil {
			return fmt.Errorf("renew job claim: %w", err)
		}
		if !ok {
			return errors.New("job claim is no longer valid")
		}
		return nil
	}
	// 认领后立即确认锁有效，后续读取和处理共用受续约控制的 Context。
	if err := renew(); err != nil {
		cancel(err)
		return nil, nil, err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("Job claim renewal panicked", "panic", recovered, "stack", string(debug.Stack()))
				cancel(fmt.Errorf("job claim renewal panicked: %v", recovered))
			}
		}()
		ticker := time.NewTicker(renewalInterval)
		defer ticker.Stop()
		for {
			select {
			case <-renewCtx.Done():
				return
			case <-ticker.C:
				if err := renew(); err != nil {
					cancel(err)
					return
				}
			}
		}
	}()
	stopped := errors.New("job claim renewal stopped")
	stopRenewal := func() error {
		cancel(stopped)
		<-done
		if cause := context.Cause(renewCtx); cause != stopped {
			return cause
		}
		return nil
	}
	return renewCtx, stopRenewal, nil
}

func (s *service) dispatchJob(ctx context.Context, job *Job) error {
	return s.store.PublishJob(ctx, job)
}
