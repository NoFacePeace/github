package executor

import "time"

// SteerPromptMode 指定 Executor 的输入模式。
type SteerPromptMode string

const (
	// SteerPromptModeLegacy 表示只从 stdin 读取一次敏感配置。
	SteerPromptModeLegacy SteerPromptMode = "legacy"
	// SteerPromptModeSteer 表示首帧读取敏感配置，后续帧持续读取控制命令。
	SteerPromptModeSteer SteerPromptMode = "steer"
)

// Spec 描述本次引擎执行的规格，具体配置字段后续补充。
type Spec struct {
	// SchemaVersion 标识进程协议版本。
	SchemaVersion string `json:"schemaVersion"`
	// ExecutionID 标识本次执行，与业务轮次标识独立。
	ExecutionID string `json:"executionId"`
	// TaskID 标识所属任务。
	TaskID string `json:"taskId"`
	// SessionID 标识所属业务轮次。
	SessionID string `json:"sessionId"`
	// Engine 指定执行引擎，例如 codex 或 claude-code。
	Engine string `json:"engine"`
	// Prompt 是本轮执行的提示词。
	Prompt string `json:"prompt"`
	// ResumeSessionID 指定待恢复的引擎会话，为空时启动新会话。
	ResumeSessionID string `json:"resumeSessionId,omitempty"`
	// SteerPromptMode 选择输入模式，为空时使用普通模式。
	SteerPromptMode SteerPromptMode `json:"steerPromptMode,omitempty"`
	// Paths 指定引擎执行使用的目录。
	Paths Paths `json:"paths"`
	// Config 保存普通引擎配置，不包含敏感凭据。
	Config Config `json:"config"`
}

// Paths 描述引擎执行使用的目录。
type Paths struct {
	// Workspace 是任务工作区目录。
	Workspace string `json:"workspace"`
	// Config 是引擎配置目录。
	Config string `json:"config"`
	// Session 是引擎会话文件和执行日志目录。
	Session string `json:"session"`
}

// Config 保存与引擎执行相关的普通配置。
type Config struct {
	// Model 指定使用的模型，为空时使用引擎默认模型。
	Model string `json:"model,omitempty"`
	// SystemPrompt 定义引擎的固定角色和规则。
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// Secrets 保存通过 stdin 管道传入的敏感执行配置。
type Secrets struct {
	// SchemaVersion 标识敏感输入协议版本。
	SchemaVersion string `json:"schemaVersion"`
	// ExecutionID 绑定本次执行，须与 Spec 中的执行标识一致。
	ExecutionID string `json:"executionId"`
	// ProviderAPIKey 是模型服务的 API Key。
	ProviderAPIKey string `json:"providerApiKey,omitempty"`
	// CodexAuthJSON 是 Codex 订阅认证信息的 JSON 文本。
	CodexAuthJSON string `json:"codexAuthJson,omitempty"`
}

// SteerCommand 表示在执行过程中追加的输入。
type SteerCommand struct {
	// CommandID 标识命令，重试时复用，供执行流程去重。
	CommandID string `json:"commandId"`
	// ExecutionID 限定命令所属的执行。
	ExecutionID string `json:"executionId"`
	// Prompt 是追加的提示词。
	Prompt string `json:"prompt"`
}

// Result 保存引擎执行的终态结果。
type Result struct {
	// ExitCode 是引擎进程退出码，nil 表示未取得退出码，0 表示正常退出。
	ExitCode *int `json:"exitCode,omitempty"`
	// ExitSignal 记录导致引擎进程退出的信号，无信号信息时为空。
	ExitSignal string `json:"exitSignal,omitempty"`
	// Text 是最终结果文本，nil 表示未提供，与已提供的空文本区分。
	Text *string `json:"text,omitempty"`
	// ExecutorSessionID 是引擎侧会话标识，供后续执行恢复上下文。
	ExecutorSessionID string `json:"executorSessionId,omitempty"`
	// Usage 是引擎报告的 Token 用量，nil 表示未提供用量信息。
	Usage *Usage `json:"usage,omitempty"`
	// Error 保存执行失败原因，成功时为空。
	Error string `json:"error,omitempty"`
}

// Usage 保存引擎报告的 Token 用量，各字段为 nil 时表示未提供该项。
type Usage struct {
	// InputTokens 是输入 Token 数。
	InputTokens *int64 `json:"inputTokens,omitempty"`
	// CachedInputTokens 是命中缓存的输入 Token 数。
	CachedInputTokens *int64 `json:"cachedInputTokens,omitempty"`
	// OutputTokens 是输出 Token 数。
	OutputTokens *int64 `json:"outputTokens,omitempty"`
	// ReasoningOutputTokens 是推理输出 Token 数。
	ReasoningOutputTokens *int64 `json:"reasoningOutputTokens,omitempty"`
	// TotalTokens 是引擎报告的总 Token 数，未提供时不自行假定计算规则。
	TotalTokens *int64 `json:"totalTokens,omitempty"`
}

// EventType 表示 Executor 进程事件类型。
type EventType string

const (
	// EventTypeStarted 表示本次执行开始。
	EventTypeStarted EventType = "started"
	// EventTypeStdout 携带引擎标准输出。
	EventTypeStdout EventType = "stdout"
	// EventTypeStderr 携带引擎标准错误输出。
	EventTypeStderr EventType = "stderr"
	// EventTypeCompleted 表示执行成功结束。
	EventTypeCompleted EventType = "completed"
	// EventTypeFailed 表示执行失败。
	EventTypeFailed EventType = "failed"
	// EventTypeStopped 表示执行已停止。
	EventTypeStopped EventType = "stopped"
)

// Event 是 Executor 通过 stdout JSONL 回传的统一进程事件。
type Event struct {
	// SchemaVersion 标识进程事件协议版本。
	SchemaVersion string `json:"schemaVersion"`
	// ExecutionID 标识事件所属的执行。
	ExecutionID string `json:"executionId"`
	// Seq 是本次执行内从 1 开始递增的事件序号。
	Seq int64 `json:"seq"`
	// Timestamp 记录事件产生时间。
	Timestamp time.Time `json:"timestamp"`
	// Type 指定事件类型。
	Type EventType `json:"type"`
	// Payload 保存输出、结果或错误等具体事件内容。
	Payload map[string]any `json:"payload,omitempty"`
}
