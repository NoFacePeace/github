// Package agentloop 提供模型与工具的执行循环。
package agentloop

import "context"

// Runner 定义执行循环的能力。
type Runner interface {
	Run(ctx context.Context) error
}

// loop 实现模型与工具的执行循环。
type loop struct{}

// New 创建执行循环实例。
func New() Runner {
	return &loop{}
}

// Run 执行模型与工具循环。
func (l *loop) Run(ctx context.Context) error {
	// TODO: 实现模型调用与工具执行循环。
	return nil
}
