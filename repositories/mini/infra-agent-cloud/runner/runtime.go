package runner

import (
	"context"
	"io"

	"infra-agent-cloud/executor"
)

// RuntimeMode 表示 Runtime 的执行模式。
type RuntimeMode string

const RuntimeModeBwrap RuntimeMode = "bwrap"

// RuntimePreflightResult 记录执行环境检查的结果。
type RuntimePreflightResult struct {
	Mode               RuntimeMode
	BwrapVersion       string
	RootlessKitVersion string
	Slirp4netnsVersion string
	TaskCgroupRoot     string
	ExecutorVersion    string
	AvailableEngines   []string
	EngineVersions     map[string]string
}

// ExecutionHandle 表示一次已启动的 Executor 执行。
type ExecutionHandle struct {
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
	// FailureReason 可选地读取隔离环境记录的失败原因。
	FailureReason func(ctx context.Context) (string, error)

	bwrapArtifacts bwrapArtifacts
	exited         <-chan struct{}
}

// Runtime 管理 Executor 进程及其隔离资源。
type Runtime interface {
	ProcessRuntime
	Mode() RuntimeMode
	// Preflight 检查运行环境和可用引擎。
	Preflight(ctx context.Context) (RuntimePreflightResult, error)
	// Recover 在节点启动时清理上次遗留的进程和隔离资源。
	Recover(ctx context.Context) error
	// ReadCodexAuthJSON 读取本轮的 Codex 认证文件，文件不存在时返回空字符串。
	ReadCodexAuthJSON(ctx context.Context, handle *ExecutionHandle) (string, error)
}

// ProcessRuntime 管理 Runner 消费一条 Job 期间的进程生命周期。
type ProcessRuntime interface {
	// Start 启动 Executor 并返回本轮句柄，secrets 可为 nil。
	// 使用隔离环境时，须在 Executor 运行前完成 cgroup 归属。
	Start(ctx context.Context, spec *executor.Spec, secrets *executor.Secrets) (*ExecutionHandle, error)
	// Stop 请求 Executor 停止，并等待进程退出。
	Stop(ctx context.Context, handle *ExecutionHandle) error
	// Destroy 清理本轮临时文件和隔离资源。
	Destroy(ctx context.Context, handle *ExecutionHandle) error
}
