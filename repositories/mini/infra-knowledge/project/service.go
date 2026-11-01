package project

import "context"

// Service 定义项目模块对外提供的业务接口。
type Service interface {
	Create(ctx context.Context, request CreateProjectRequest) (Project, error)
}

// CreateProjectRequest 表示创建项目的请求。
type CreateProjectRequest struct {
	Name string
}
