package agentloop

import "context"

// Recorder 定义执行循环的过程记录能力。
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}

// EntryType 表示执行循环产生的记录类型。
type EntryType string

const (
	EntryTypePaused         EntryType = "paused"
	EntryTypeContextTrimmed EntryType = "context_trimmed"
)

// Entry 描述一条执行记录，由记录器补充关联信息并保存。
type Entry struct {
	Type EntryType
	Data map[string]any
}
