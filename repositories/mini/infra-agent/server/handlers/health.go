// Package handlers 提供 HTTP 请求处理器。
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"infra-agent/server/services"
)

// HealthHandler 处理健康检查请求。
type HealthHandler struct {
	healthService *services.HealthService
}

// NewHealthHandler 创建健康检查处理器。
func NewHealthHandler(healthService *services.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// Ping 返回服务连通性检查结果。
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": h.healthService.Ping(c.Request.Context())})
}
