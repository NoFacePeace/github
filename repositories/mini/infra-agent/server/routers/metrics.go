package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	routeRegistrars = append(routeRegistrars, registerMetricsRoutes)
}

// registerMetricsRoutes 注册 Prometheus 指标接口。
func registerMetricsRoutes(router gin.IRouter, _ Dependencies) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
