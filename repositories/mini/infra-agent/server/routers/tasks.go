package routers

import "github.com/gin-gonic/gin"

func init() {
	routeRegistrars = append(routeRegistrars, registerTaskRoutes)
}

// registerTaskRoutes 注册任务接口。
func registerTaskRoutes(router gin.IRouter, dependencies Dependencies) {
	router.POST("/tasks", dependencies.TaskHandler.Create)
}
