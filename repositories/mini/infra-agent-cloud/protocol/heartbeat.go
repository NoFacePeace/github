package protocol

import "time"

// Heartbeat 描述节点身份、引擎能力和并发上限。
type Heartbeat struct {
	// RunnerID 使用 Server 当前读取的 ID 字段名。
	RunnerID string `json:"ID"`
	// InstanceID 标识一次 Runner 进程启动。
	InstanceID string
	Status     string
	Engines    []string
	Capacity   int
	ExpiresAt  time.Time
}
