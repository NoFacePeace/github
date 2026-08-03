// Package routers 提供 HTTP 路由的集中注册入口。
package routers

import (
	"github.com/gin-gonic/gin"
)

var routeRegistrars []func(gin.IRouter, Dependencies)

// Register 将应用路由注册到 Gin 引擎。
func Register(router *gin.Engine, dependencies Dependencies) {
	for _, registerRoutes := range routeRegistrars {
		registerRoutes(router, dependencies)
	}
}
