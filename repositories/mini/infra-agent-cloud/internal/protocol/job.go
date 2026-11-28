package protocol

import "time"

// Job 是 Server 投递给指定 Runner 的执行消息。
type Job struct {
	MessageID         string    `json:"message_id"`
	TaskID            string    `json:"task_id"`
	SessionID         string    `json:"session_id"`
	RunnerID          string    `json:"runner_id"`
	ExecutorSessionID string    `json:"executor_session_id"`
	Prompt            string    `json:"prompt"`
	AgentVersionID    string    `json:"agent_version_id"`
	Engine            string    `json:"engine"`
	Model             string    `json:"model"`
	SystemPrompt      string    `json:"system_prompt"`
	LeaseID           string    `json:"lease_id"`
	LeaseExpiresAt    time.Time `json:"lease_expires_at"`
}
