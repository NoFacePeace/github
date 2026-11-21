package watch

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handle 建立 Watch 响应流，并保持连接直到请求取消。
func Handle(c *gin.Context) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Status(http.StatusOK)
	c.Writer.Flush()
	<-c.Request.Context().Done()
}
