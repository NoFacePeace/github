package agentloop

// EntryType 表示执行循环产生的记录类型。
type EntryType string

// Entry 描述一条执行记录，由记录器补充关联信息并保存。
type Entry struct {
	Type EntryType
	Data map[string]any
}
