// Package providers 提供模型供应商相关的领域模型。
package providers

import "context"

// Provider 表示一个模型供应商。
type Provider struct {
	ID   string
	Name string
}

// Manager 定义供应商管理接口。
type Manager interface {
	// Get 根据 ID 获取供应商。
	Get(ctx context.Context, id string) (Provider, error)
}
