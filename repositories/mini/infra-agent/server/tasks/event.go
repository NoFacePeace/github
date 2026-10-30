package tasks

import (
	"encoding/json"
	"time"
)

// EventType 表示任务运行过程中发生的事件类型。
type EventType string

const (
	// EventTaskCreated 表示任务已创建。
	EventTaskCreated EventType = "task_created"
	// EventMessageReceived 表示收到用户消息。
	EventMessageReceived EventType = "message_received"
)

// Event 表示一条关联任务和运行的业务事件。
type Event struct {
	ID        string
	TaskID    string
	RunID     string
	Type      EventType
	Actor     string
	Payload   json.RawMessage
	Seq       int64
	CreatedAt time.Time
}
