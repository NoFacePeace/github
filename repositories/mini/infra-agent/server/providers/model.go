package providers

import "context"

// Model 表示供应商下的一条模型配置。
type Model struct {
	ID              string // 本地模型配置记录的 ID。
	ProviderID      string
	Name            string
	ModelID         string // 调用供应商 API 时使用的模型标识。
	Enabled         bool   // 是否启用。
	MaxOutputTokens int    // 最大输出 token 数，0 表示未配置。
}

// ModelManager 定义模型管理接口。
type ModelManager interface {
	// List 返回全部模型，包括未启用的模型。
	List(ctx context.Context) ([]Model, error)
}
