package server

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
}
