package executor

import (
	"context"
	"errors"
)

type Codex struct{}

func NewCodex() *Codex {
	return &Codex{}
}

func (c *Codex) Run(ctx context.Context, spec *ExecutionSpec, emit func(Event) error) (*Result, error) {
	// TODO: 启动 Codex、解析流式输出，并在取消时终止进程。
	return nil, errors.New("codex execution is not implemented")
}
