package providers

// Model 表示供应商下的一条模型配置。
type Model struct {
	ID         string // 本地模型配置记录的 ID。
	ProviderID string
	Name       string
	ModelID    string // 调用供应商 API 时使用的模型标识。
}

// ModelManager 定义模型管理接口。
type ModelManager interface {
}
