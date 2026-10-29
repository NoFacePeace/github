package workspaces

import "context"

// Service 定义工作区模块对外提供的业务接口。
type Service interface {
	Get(ctx context.Context, id string) (Workspace, error)
}
