package model

import "context"

type Model interface {
	// Generate 按顺序串行调用 onEvent，成功时最后发送一次 CompletedEvent。
	// 返回前必须结束所有事件回调；CompletedEvent.Message 包含完整文本和工具调用。
	// 实现负责响应 ctx 的取消和超时，并返回对应错误。
	Generate(ctx context.Context, req Request, onEvent func(Event)) error
}
