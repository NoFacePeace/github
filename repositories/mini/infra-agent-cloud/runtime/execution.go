package runtime

import (
	"context"
	"io"
)

// Execution 表示一次已启动的 Executor 执行。
type Execution struct {
	ExecutionID          string
	PID                  int
	InterruptPID         int
	InterruptProcessOnly bool
	Stdout               io.Reader
	Stderr               io.Reader

	// SendInput 仅在本轮保留控制输入管道时可用，调用方提供包含换行符的完整命令帧。
	SendInput func(ctx context.Context, input string) error
	// Wait 等待进程退出及调用方读完或关闭两路输出，返回退出码。
	Wait func(ctx context.Context) (int, error)
	// WaitForExit 等待进程退出，返回退出码。
	WaitForExit func(ctx context.Context) (int, error)
	// FailureReason 可选地读取运行环境记录的失败原因。
	FailureReason func(ctx context.Context) (string, error)
}
