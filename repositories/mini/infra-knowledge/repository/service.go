package repository

import "context"

// Service 定义仓库模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateRepositoryRequest) (Repository, error)
}

// CreateRepositoryRequest 表示创建代码仓库记录的请求。
type CreateRepositoryRequest struct {
	Name      string
	GitURL    string
	ProjectID string
}
