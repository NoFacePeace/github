package main

import (
	"net/http"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/gu"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/tencent/gu/hkminute", func(c *gin.Context) {
		code := c.Query("code")
		ps, err := gu.HKMinute(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ps)
	})
	r.Run()
}
