package config

// Config 定义多个 Controller 组件共用的配置。
type Config struct {
	// MiddlewareType 表示当前管理的中间件类型。
	MiddlewareType string
	// ClusterFilter 定义需要管理的集群范围。
	ClusterFilter ClusterFilter
}

// ClusterFilter 定义集群白名单与黑名单，匹配逻辑后续补充。
type ClusterFilter struct {
	// Allowlist 保存允许管理的集群标识。
	Allowlist []string
	// Denylist 保存排除管理的集群标识。
	Denylist []string
}
