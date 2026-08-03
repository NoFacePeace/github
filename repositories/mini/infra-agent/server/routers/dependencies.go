package routers

import "infra-agent/server/handlers"

// Dependencies 保存路由注册所需的处理器依赖。
type Dependencies struct {
	HealthHandler *handlers.HealthHandler
	TaskHandler   *handlers.TaskHandler
}
