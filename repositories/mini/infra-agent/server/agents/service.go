package agents

import "context"

// Service 定义智能体模块对外提供的业务接口。
type Service interface {
	Get(ctx context.Context, id string) (Agent, error)
}
