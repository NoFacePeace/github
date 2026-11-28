package runner

import (
	"context"
	"time"

	"infra-agent-cloud/executor"
)

// Store 封装 Runner 的 Redis 心跳、消息消费和事件回传。
type Store interface {
	// PublishHeartbeat 上报节点信息并刷新在线有效期，首次上报同时完成注册。
	PublishHeartbeat(ctx context.Context, heartbeat Heartbeat) error
	// ClaimJob 从当前 Runner 的执行通道返回任务和 Redis 消息 ID，无任务时返回 nil, "", nil。
	// 此处领取的是 Server 已分配的执行消息，不是 Server 的待调度队列。
	ClaimJob(ctx context.Context) (job *Job, entryID string, err error)
	// AckJob 在处理 Job 后确认消费，包括处理或失败事件回传报错的情况。重复确认须幂等。
	AckJob(ctx context.Context, entryID string) error
	// ClaimCommand 领取当前节点的控制命令，无命令时返回 nil, nil。
	ClaimCommand(ctx context.Context) (*Command, error)
	AckCommand(ctx context.Context, command *Command) error
	NackCommand(ctx context.Context, command *Command, reason string) error
	// PublishEvent 可靠回传执行事件；具体事件标识、排序和幂等协议后续补充。
	PublishEvent(ctx context.Context, job *Job, event Event) error
}

// Heartbeat 描述节点身份、引擎能力和并发上限。
type Heartbeat struct {
	RunnerID string
	Engines  []string
	Capacity int
}

type CommandType string

const (
	CommandTypeStop CommandType = "stop"
)

// Command 对应 Server 投递给指定 Runner 的控制消息。
type Command struct {
	ID             string
	Type           CommandType
	RunnerID       string
	TaskID         string
	SessionID      string
	ExecutionToken string
	ExpiresAt      time.Time
}

type EventType string

const (
	EventTypeStarted   EventType = "started"
	EventTypeOutput    EventType = "output"
	EventTypeCompleted EventType = "completed"
	EventTypeFailed    EventType = "failed"
	EventTypeStopped   EventType = "stopped"
)

// Event 是向 Server 回传的执行事件，关联信息由 Job 提供。
type Event struct {
	Type   EventType
	Error  string
	Output *executor.Event
	Result *executor.Result
}
