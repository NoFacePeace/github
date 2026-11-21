package apiserver

import (
	"umo/apiserver/watch"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建 API Server 的 Gin 路由。
func NewRouter() *gin.Engine {
	router := gin.New()
	router.GET("/watch/:clusterID", watch.Handle)
	return router
}
