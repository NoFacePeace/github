package server

import "time"

type Task struct {
	ID             string
	AgentID        string
	AgentVersionID string
}

type SessionStatus string

const (
	// SessionStatusRunning 覆盖排队、调度和执行阶段，表示本轮会话尚未结束。
	SessionStatusRunning   SessionStatus = "running"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
	SessionStatusStopped   SessionStatus = "stopped"
)

type Session struct {
	ID       string
	TaskID   string
	RunnerID string
	// ExecutorSessionID 是引擎侧的会话标识，继续任务时复用以恢复上下文。
	ExecutorSessionID string
	Prompt            string
	Status            SessionStatus
	Error             string
	// CreatedAt 用于恢复扫描判断会话是否仍处于启动宽限期。
	CreatedAt time.Time
	// Dispatch 保存调度状态摘要，不保存提示词、执行租约或完整 Job。
	Dispatch *SessionDispatch
}

// SessionDispatch 描述本轮会话的调度进度和结果。
type SessionDispatch struct {
	MessageID string
	Status    JobStatus
	QueuedAt  time.Time
	Deadline  time.Time
	Error     string
}
