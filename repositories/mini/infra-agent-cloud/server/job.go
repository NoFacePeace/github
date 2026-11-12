package server

import (
	"errors"
	"time"
)

type JobStatus string

const (
	JobStatusQueued      JobStatus = "queued"
	JobStatusDispatching JobStatus = "dispatching"
	JobStatusDispatched  JobStatus = "dispatched"
	JobStatusFailed      JobStatus = "failed"
	JobStatusCancelled   JobStatus = "cancelled"
)

var (
	// ErrExecutionLeaseExpired 表示未发布消息携带的执行租约已到期，不能继续投递。
	ErrExecutionLeaseExpired = errors.New("job execution lease expired")
	// ErrDispatchConflict 表示同一 Runner、Session 的发布标记已对应另一条消息。
	ErrDispatchConflict = errors.New("job dispatch conflict")
)

type Job struct {
	// MessageID 标识固定的执行消息，投递重试时必须复用，不得重新生成。
	MessageID string
	Status    JobStatus
	Error     string
	QueuedAt  time.Time
	TaskID    string
	SessionID string
	RunnerID  string
	// ExecutorSessionID 为空时启动新引擎会话，非空时恢复已有会话。
	ExecutorSessionID string
	Prompt            string
	AgentVersionID    string
	Engine            string
	Model             string
	SystemPrompt      string
	ClaimToken        string
	ClaimExpiresAt    time.Time
	// ExecutionToken 用于校验执行租约的续约和释放权限，与调度领取 token 独立。
	ExecutionToken string
	// ExecutionExpiresAt 表示执行租约的过期时间，用于 Runner 容量占用和执行失联判断。
	ExecutionExpiresAt time.Time
	// Deadline 表示排队调度的截止时间，不是 Runner 执行超时。
	Deadline time.Time
}

// ExecutionLease 保存本次执行的归属和有效期，与 Job 分 key 存储。
type ExecutionLease struct {
	TaskID             string
	SessionID          string
	RunnerID           string
	ExecutionToken     string
	ExecutionExpiresAt time.Time
	// LastSeq 是最后接受的执行事件序号。
	LastSeq int64
}
