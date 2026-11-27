package runner

import "time"

// Job 对应 Server 投递给指定 Runner 的执行消息。
type Job struct {
	MessageID          string
	TaskID             string
	SessionID          string
	RunnerID           string
	ExecutorSessionID  string
	Prompt             string
	AgentVersionID     string
	Engine             string
	Model              string
	SystemPrompt       string
	ExecutionToken     string
	ExecutionExpiresAt time.Time
}
