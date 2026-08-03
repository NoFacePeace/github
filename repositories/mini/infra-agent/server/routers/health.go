package routers

import "github.com/gin-gonic/gin"

func init() {
	routeRegistrars = append(routeRegistrars, registerHealthRoutes)
}

// registerHealthRoutes 注册健康检查接口。
func registerHealthRoutes(router gin.IRouter, dependencies Dependencies) {
	router.GET("/ping", dependencies.HealthHandler.Ping)
}
