package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建 HTTP 路由，业务模块不依赖 Gin。
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}
