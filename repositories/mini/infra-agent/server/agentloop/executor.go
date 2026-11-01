package agentloop

import (
	"context"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// ToolExecutor 定义工具执行能力。
type ToolExecutor interface {
	Execute(ctx context.Context, calls []llm.ToolCall) ([]tools.Result, error)
}
