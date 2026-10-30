// Package agents 提供智能体相关的领域模型和业务接口。
package agents

// DefaultAgentID 是未指定智能体时使用的默认 ID。
const DefaultAgentID = "default"

// Agent 表示一个智能体。
type Agent struct {
	ID string
}
