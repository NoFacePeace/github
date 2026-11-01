// Package llm 提供与大模型交互的通用抽象。
package llm

import (
	"context"
	"encoding/json"
)

// Model 定义大模型模块对外提供的调用接口。
type Model interface {
	Generate(ctx context.Context, request Request) (Response, error)
}

// Request 表示一次大模型调用请求。
type Request struct {
	Messages []Message
	Tools    []ToolDefinition
}

// ToolDefinition 描述模型可调用的工具及其参数结构。
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// Role 表示消息的角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message 表示模型上下文中的一条消息。
type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

// ToolCall 表示模型请求的一次工具调用。
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// Response 表示一次大模型调用结果。
type Response struct {
	Content string
}
