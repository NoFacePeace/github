package model

import "github.com/NoFacePeace/github/repositories/go/demo/agent/tool"

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []tool.Call
	ToolCallID string // 工具结果所对应的调用 ID
}
