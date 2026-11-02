package tasks

// AuthorizationScope 表示工具授权的适用范围。
type AuthorizationScope string

const (
	AuthorizationScopeOnce      AuthorizationScope = "once"
	AuthorizationScopeRun       AuthorizationScope = "run"
	AuthorizationScopeWorkspace AuthorizationScope = "workspace"
)

// ToolAuthorization 表示已批准的工具授权，不包含联网和路径访问授权。
type ToolAuthorization struct {
	ToolName string
	// Key 对应工具返回的授权匹配键。
	Key   string
	Scope AuthorizationScope
	// ToolCallID 和 RunID 共同限定单次调用授权。
	ToolCallID string
	// RunID 限定单次调用和运行级授权。
	RunID string
	// WorkspaceID 限定工作区级授权。
	WorkspaceID string
}

// NetworkAuthorization 表示当前运行已获准联网，不区分工具和网络目标。
type NetworkAuthorization struct {
	RunID string
}

// PathAuthorization 表示指定工作区中可复用的路径访问授权。
type PathAuthorization struct {
	WorkspaceID string
	// Path 是规范化的绝对路径，目录授权覆盖其子路径。
	Path string
	// Read 和 Write 分别表示允许读取和写入，可以同时授予。
	Read  bool
	Write bool
}
