package model

import "github.com/NoFacePeace/github/repositories/go/demo/agent/tool"

type Event interface{}

// TextDeltaEvent 表示新增的文本片段。
type TextDeltaEvent struct {
	Text string
}

// ToolCallEvent 表示参数已经完整、可以执行的工具调用。
type ToolCallEvent struct {
	Call tool.Call
}

// CompletedEvent 表示本轮模型生成结束，并携带完整回复。
type CompletedEvent struct {
	Message Message
}
