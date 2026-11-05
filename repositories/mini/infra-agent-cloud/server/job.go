package server

import "time"

type JobStatus string

const (
	JobStatusQueued      JobStatus = "queued"
	JobStatusDispatching JobStatus = "dispatching"
	JobStatusDispatched  JobStatus = "dispatched"
	JobStatusFailed      JobStatus = "failed"
	JobStatusCancelled   JobStatus = "cancelled"
)

type Job struct {
	Status         JobStatus
	Error          string
	QueuedAt       time.Time
	TaskID         string
	SessionID      string
	Prompt         string
	AgentVersionID string
	Engine         string
	Model          string
	SystemPrompt   string
	ClaimToken     string
	ClaimExpiresAt time.Time
	// Deadline 表示排队调度的截止时间，不是 Runner 执行超时。
	Deadline time.Time
	// NextAttemptAt 表示下次允许领取的时间，零值表示无需等待。
	NextAttemptAt time.Time
}
