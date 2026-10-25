package model

type Event interface{}

// TextDeltaEvent 表示新增的文本片段。
type TextDeltaEvent struct {
	Text string
}

// CompletedEvent 表示本轮模型生成结束，并携带完整回复。
type CompletedEvent struct {
	Message Message
}
