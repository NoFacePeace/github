package loop

import "github.com/NoFacePeace/github/repositories/go/demo/agent/tool"

type Event interface{}

// TextDeltaEvent 表示模型新增的文本片段。
type TextDeltaEvent struct {
	Text string
}

// ToolCallEvent 表示开始执行工具调用。
type ToolCallEvent struct {
	Call tool.Call
}

// ToolResultEvent 表示一次工具调用的执行结果。
type ToolResultEvent struct {
	CallID string
	Result tool.Result
}

// CompletedEvent 表示整个任务结束。
type CompletedEvent struct {
	Reason StopReason
}

type StopReason string

const (
	StopCompleted StopReason = "completed"
	StopMaxSteps  StopReason = "max_steps"
)
