package server

type Task struct {
	ID             string
	AgentID        string
	AgentVersionID string
}

type SessionStatus string

const (
	SessionStatusPending   SessionStatus = "pending"
	SessionStatusRunning   SessionStatus = "running"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
	SessionStatusStopped   SessionStatus = "stopped"
)

type Session struct {
	ID     string
	TaskID string
	Prompt string
	Status SessionStatus
	Error  string
}
