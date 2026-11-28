package runtime

import (
	"context"

	"infra-agent-cloud/executor"
)

// Runtime 管理 Executor 进程及其运行环境。
type Runtime interface {
	// Start 启动 Executor 并返回本次执行的句柄，secrets 可为 nil。
	Start(ctx context.Context, spec *executor.Spec, secrets *executor.Secrets) (*Execution, error)
	// Stop 请求停止指定的执行，并等待进程退出。
	Stop(ctx context.Context, execution *Execution) error
	// Recover 在 Runner 启动时处理遗留的进程和运行环境资源。
	Recover(ctx context.Context) error
	// Destroy 在进程退出后清理本次执行的临时资源。
	Destroy(ctx context.Context, execution *Execution) error
}
