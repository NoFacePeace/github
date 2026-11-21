package apiserver

import (
	"umo/apiserver/watch"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建 API Server 的 Gin 路由。
func NewRouter(store watch.Store, broadcaster *watch.EventBroadcaster) *gin.Engine {
	router := gin.New()
	router.GET("/watch/:clusterID", watch.NewHandler(store, broadcaster))
	return router
}
