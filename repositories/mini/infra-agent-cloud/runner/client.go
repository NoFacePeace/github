package runner

import (
	"context"
	"infra-agent-cloud/executor"
	"time"
)

// Client 封装 Runner 与服务端的通信。
type Client interface {
	// PublishHeartbeat 上报节点信息并刷新在线有效期，首次上报同时完成注册。
	PublishHeartbeat(ctx context.Context, heartbeat Heartbeat) error
	// ClaimJob 从当前 Runner 的执行通道领取任务，无任务时返回 nil, nil。
	// 此处领取的是 Server 已分配的执行消息，不是 Server 的待调度队列。
	ClaimJob(ctx context.Context) (*Job, error)
	// AckJob 在终态已可靠回传后确认消费；重复确认须幂等。
	AckJob(ctx context.Context, job *Job) error
	// NackJob 在处理或事件回传失败时保留消息供恢复，不能因此重复启动已执行的任务。
	NackJob(ctx context.Context, job *Job, reason string) error
	// ClaimCommand 领取当前节点的控制命令，无命令时返回 nil, nil。
	ClaimCommand(ctx context.Context) (*Command, error)
	AckCommand(ctx context.Context, command *Command) error
	NackCommand(ctx context.Context, command *Command, reason string) error
	// RenewExecution 向 Server 请求执行续约，返回确认的过期时间；须区分授权拒绝与临时通信错误。
	RenewExecution(ctx context.Context, job *Job) (time.Time, error)
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

// Command 是 Runner 接收的控制消息，目标执行 token、命令 ID 和有效期后续补充。
type Command struct {
	Type CommandType
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
