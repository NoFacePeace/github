package executor

import (
	"context"
	"io"
)

// Engine 定义 Executor 使用的引擎适配器接口。
type Engine interface {
	Name() string
	WriteConfig(spec *Spec, secrets *Secrets) error
	BuildCommand(spec *Spec, secrets *Secrets) (EngineCommand, error)
	// StartSteer 接管引擎的原始 stdout，初始化交互协议并发送首条提示词。
	// 轮次结束时，由适配器结束待回复的请求，并在引擎需要 EOF 时关闭 stdin。
	StartSteer(ctx context.Context, spec *Spec, stdin io.WriteCloser, stdout io.ReadCloser) (SteerSession, error)

	// ParseSessionID 提取引擎会话标识，当前行没有会话标识时返回空字符串。
	ParseSessionID(line string) string
	// ParseMCPStatus 提取 MCP 服务和工具状态，没有状态时返回 nil。
	ParseMCPStatus(line string) *EngineMCPStatus
	// ParseProgress 提取实时进度，没有进度时返回 nil。
	ParseProgress(line string) *EngineProgress
	// ParseChannelOutput 提取当前行可展示的进度或回答。
	ParseChannelOutput(line string) []EngineChannelOutput
	// ParseResult 在进程退出后解析完整的原始标准输出。
	ParseResult(log string, spec *Spec) (EngineResult, error)
}

// EngineLogFilter 可选地处理引擎标准输出的日志副本，返回 nil 时跳过该行。
type EngineLogFilter interface {
	FilterLogLine(line string) *string
}

// EngineErrorLogFilter 可选地处理引擎标准错误的日志副本，返回 nil 时跳过该行。
type EngineErrorLogFilter interface {
	FilterErrorLine(line string) *string
}

// EngineCommand 描述引擎的启动命令，Env 保存需要覆盖的环境变量。
type EngineCommand struct {
	Executable string
	Args       []string
	Dir        string
	Env        []string
}

// EngineProgress 表示一条进度，Replace 表示替换上一条进度。
type EngineProgress struct {
	Text    string
	Replace bool
}

// EngineChannelOutputKind 区分面向用户展示的进度和回答。
type EngineChannelOutputKind string

const (
	EngineChannelOutputProgress EngineChannelOutputKind = "progress"
	EngineChannelOutputAnswer   EngineChannelOutputKind = "answer"
)

// EngineChannelOutput 表示一条可展示的引擎输出。
type EngineChannelOutput struct {
	EngineProgress
	Kind EngineChannelOutputKind
}

// EngineMCPStatus 描述引擎报告的 MCP 服务连接状态和工具列表。
type EngineMCPStatus struct {
	SessionID     string                  `json:"sessionId,omitempty"`
	Healthy       bool                    `json:"healthy"`
	Servers       []EngineMCPServerStatus `json:"servers"`
	FailedServers []string                `json:"failedServers"`
	Tools         []string                `json:"tools"`
}

// EngineMCPServerStatus 描述单个 MCP 服务的状态。
type EngineMCPServerStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// EngineResult 保存从引擎输出解析出的终态信息。
type EngineResult struct {
	Text          *string
	Usage         *Usage
	CompactCount  *int
	Cost          *float64
	CostEstimated *bool
}
