package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/signal"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := signal.SetupSignalHandler()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	srv := &http.Server{
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("http server listen and server error", "error", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "error", err)
		os.Exit(1)
	}
}
