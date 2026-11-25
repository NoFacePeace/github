package executor

import (
	"fmt"
	"sync"
	"time"
)

// EventType 表示 Executor 进程事件类型。
type EventType string

const (
	// EventTypeStarted 表示本次执行开始。
	EventTypeStarted EventType = "started"
	// EventTypeExecutorReady 表示引擎已准备好接收后续 steer 命令。
	EventTypeExecutorReady EventType = "executor_ready"
	// EventTypeSteerResult 携带单条 steer 命令的处理结果。
	EventTypeSteerResult EventType = "steer_result"
	// EventTypeStdout 携带引擎标准输出。
	EventTypeStdout EventType = "stdout"
	// EventTypeStderr 携带引擎标准错误输出。
	EventTypeStderr EventType = "stderr"
	// EventTypeChannelOutput 携带可展示的进度或回答，由 kind 区分。
	EventTypeChannelOutput EventType = "channel_output"
	// EventTypeCompleted 表示执行成功结束。
	EventTypeCompleted EventType = "completed"
	// EventTypeFailed 表示执行失败。
	EventTypeFailed EventType = "failed"
	// EventTypeStopped 表示执行已停止。
	EventTypeStopped EventType = "stopped"
)

// Event 是 Executor 通过 stdout JSONL 回传的统一进程事件。
type Event struct {
	// SchemaVersion 标识进程事件协议版本。
	SchemaVersion string `json:"schemaVersion"`
	// ExecutionID 标识事件所属的执行。
	ExecutionID string `json:"executionId"`
	// Seq 是本次执行内从 1 开始递增的事件序号。
	Seq int64 `json:"seq"`
	// Timestamp 记录事件产生时间。
	Timestamp time.Time `json:"timestamp"`
	// Type 指定事件类型。
	Type EventType `json:"type"`
	// Payload 保存输出、结果或错误等具体事件内容。
	Payload map[string]any `json:"payload,omitempty"`
}

// eventPublisher 为单次执行组装事件，并按序号串行写出。
type eventPublisher struct {
	schemaVersion string
	executionID   string
	writeEvent    func(Event) error

	mu       sync.Mutex
	seq      int64
	writeErr error
}

func (p *eventPublisher) publish(kind EventType, payload map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.writeErr != nil {
		return p.writeErr
	}
	p.seq++
	if err := p.writeEvent(Event{
		SchemaVersion: p.schemaVersion,
		ExecutionID:   p.executionID,
		Seq:           p.seq,
		Timestamp:     time.Now().UTC(),
		Type:          kind,
		Payload:       payload,
	}); err != nil {
		p.writeErr = fmt.Errorf("emit %s event: %w", kind, err)
	}
	return p.writeErr
}
