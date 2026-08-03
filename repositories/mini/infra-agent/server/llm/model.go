// Package llm 提供与大模型交互的通用抽象。
package llm

import "context"

// Model 定义大模型模块对外提供的调用接口。
type Model interface {
	Generate(ctx context.Context, request Request) (Response, error)
}

// Request 表示一次大模型调用请求。
type Request struct {
	Messages []Message
}

// Message 表示模型上下文中的一条消息。
type Message struct {
	Role    string
	Content string
}

// Response 表示一次大模型调用结果。
type Response struct {
	Content string
}
