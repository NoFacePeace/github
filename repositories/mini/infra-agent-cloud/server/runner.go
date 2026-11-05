package server

import "time"

// RunnerStatus 表示执行节点的在线状态。
type RunnerStatus string

const (
	// RunnerStatusOnline 表示节点在线，调度时仍需检查引擎能力和剩余容量。
	RunnerStatusOnline RunnerStatus = "online"
	// RunnerStatusOffline 表示节点离线，不应向其分配新的执行任务。
	RunnerStatusOffline RunnerStatus = "offline"
)

// Runner 保存调度所需的执行节点状态、引擎能力和并发容量信息。
type Runner struct {
	// ID 是执行节点的唯一标识。
	ID string
	// Status 是节点当前的在线状态。
	Status RunnerStatus
	// Engines 列出节点支持的执行引擎，例如 codex。
	Engines []string
	// Capacity 是节点允许同时执行的会话数上限。
	Capacity int
	// ActiveSessions 是当前占用的并发名额，包含已预留但尚未启动的执行。
	// 选择节点后仍需原子预留容量，不能仅凭此字段判断就直接投递。
	ActiveSessions int
}

type RunnerCommandType string

const (
	RunnerCommandTypeStop RunnerCommandType = "stop"
)

// RunnerCommand 表示发送到指定 Runner 的执行控制指令。
type RunnerCommand struct {
	// ID 标识同一条命令，重试发送时复用。
	ID        string
	Type      RunnerCommandType
	RunnerID  string
	TaskID    string
	SessionID string
	// ExecutionToken 限定目标执行，防止迟到命令误停其他执行。
	ExecutionToken string
	// ExpiresAt 表示命令有效期，Runner 不处理已过期的命令。
	ExpiresAt time.Time
}
