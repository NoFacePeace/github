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
		ID:     rand.Text(),
		TaskID: task.ID,
		Prompt: input.Prompt,
		Status: SessionStatusPending,
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
	if lastSession != nil && (lastSession.Status == SessionStatusPending || lastSession.Status == SessionStatusRunning) {
		return nil, errors.New("task already has an active session")
	}
	// 继续任务始终沿用 Task 绑定的版本，不读取 Agent 当前发布版本或草稿。
	version, err := s.store.GetVersion(ctx, task.AgentVersionID)
	if err != nil {
		return nil, err
	}
	session := &Session{
		ID:     rand.Text(),
		TaskID: task.ID,
		Prompt: input.Prompt,
		Status: SessionStatusPending,
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
		failErr := s.store.FailSession(cleanupCtx, session.ID, err.Error())
		if failErr != nil {
			return errors.Join(err, fmt.Errorf("fail session after job submission error: %w", failErr))
		}
		return err
	}
	s.wakeScheduler()
	return nil
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
		}
		if err := s.dispatchPending(ctx); err != nil {
			slog.Error("Failed to dispatch pending jobs", "error", err)
		}
	}
}

func (s *service) dispatchPending(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		job, err := s.store.ClaimJob(ctx)
		if err != nil {
			return err
		}
		if job == nil {
			return nil
		}
		if !job.Deadline.IsZero() && !time.Now().Before(job.Deadline) {
			if _, err := s.store.FailJob(ctx, job.SessionID, job.ClaimToken, "job dispatch deadline exceeded"); err != nil {
				return err
			}
			continue
		}
		if err := s.dispatchWithClaimRenewal(ctx, job); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			now := time.Now()
			if !job.Deadline.IsZero() && !now.Before(job.Deadline) {
				if _, failErr := s.store.FailJob(ctx, job.SessionID, job.ClaimToken, "job dispatch deadline exceeded"); failErr != nil {
					return failErr
				}
				continue
			}
			// 暂按可重试错误处理，后续补充确定失败与投递结果不确定的分类。
			retried, retryErr := s.store.RetryJob(ctx, job.SessionID, job.ClaimToken, err.Error(), now.Add(30*time.Second))
			if retryErr != nil {
				return retryErr
			}
			slog.Warn("Job dispatch failed", "sessionID", job.SessionID, "retryScheduled", retried, "error", err)
		}
	}
}

func (s *service) dispatchWithClaimRenewal(ctx context.Context, job *Job) error {
	const claimTTL = 30 * time.Second
	const renewalInterval = 10 * time.Second
	if job.ClaimToken == "" {
		return errors.New("job claim token is required")
	}

	dispatchCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	renew := func() error {
		renewCtx, stop := context.WithTimeout(dispatchCtx, renewalInterval)
		defer stop()
		ok, err := s.store.RenewJobClaim(renewCtx, job.SessionID, job.ClaimToken, claimTTL)
		if err != nil {
			return fmt.Errorf("renew job claim: %w", err)
		}
		if !ok {
			return errors.New("job claim is no longer valid")
		}
		return nil
	}
	// 分发前先确认领取有效，并统一延长到续约周期所需的期限。
	if err := renew(); err != nil {
		return err
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
			case <-dispatchCtx.Done():
				return
			case <-ticker.C:
				if err := renew(); err != nil {
					cancel(err)
					return
				}
			}
		}
	}()
	defer func() {
		cancel(nil)
		<-done
	}()

	err := s.dispatchJob(dispatchCtx, job)
	return errors.Join(err, context.Cause(dispatchCtx))
}

func (s *service) dispatchJob(ctx context.Context, job *Job) error {
	if job.RunnerID == "" {
		runners, err := s.store.ListRunners(ctx)
		if err != nil {
			return err
		}
		for _, runner := range runners {
			if runner.Status != RunnerStatusOnline ||
				!slices.Contains(runner.Engines, job.Engine) ||
				runner.ActiveSessions >= runner.Capacity {
				continue
			}
			job.RunnerID = runner.ID
			break
		}
		if job.RunnerID == "" {
			return errors.New("no available compatible runner")
		}
	}
	reserved, err := s.store.ReserveExecution(ctx, job.SessionID, job.ClaimToken, job.RunnerID, 5*time.Minute)
	if err != nil {
		return err
	}
	*job = *reserved
	if err := s.store.PublishJob(ctx, job); err != nil {
		return err
	}
	completed, err := s.store.CompleteJobDispatch(ctx, job.SessionID, job.ClaimToken)
	if err != nil {
		return err
	}
	if !completed {
		return errors.New("job dispatch confirmation rejected")
	}
	return nil
}
