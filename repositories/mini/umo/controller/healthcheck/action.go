package healthcheck

// Action 表示健康检查命中策略后请求的处理动作。
type Action string

// 健康检查支持的处理动作。
const (
	ActionMigrate       Action = "Migrate"
	ActionReCreate      Action = "ReCreate"
	ActionInPlaceUpdate Action = "InPlaceUpdate"
	ActionManual        Action = "Manual"
	ActionAutomatic     Action = "Automatic"
)
