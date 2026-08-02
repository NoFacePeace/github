package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginotel "github.com/NoFacePeace/github/repositories/go/utils/gin/otel"
	"github.com/NoFacePeace/github/repositories/go/utils/otel"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

func main() {
	// 监听进程退出信号，供服务和关闭流程共同使用。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done() // 第一次 Ctrl-C 或 SIGTERM：开始优雅关闭。
		stop()       // 恢复默认信号行为，第二次 Ctrl-C 会立即终止进程。
	}()

	// 初始化全局 OpenTelemetry Provider，并在进程退出时刷出未发送的遥测数据。
	shutdownOtel, err := otel.Setup(
		ctx,
		otel.WithServiceName("infra-agent"),
		otel.WithOtlptracehttp(),
		otel.WithOtlploghttp(otlploghttp.WithEndpointURL("http://localhost:3100/otlp/v1/logs")),
		otel.WithMetricPrometheus(),
	)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		// OTel Provider 尚未可用，初始化失败时使用标准日志作为兜底。
		log.Fatalf("initialize OpenTelemetry: %v", err)
	}
	serviceLogger := otelslog.NewLogger("infra-agent")
	defer func() {
		// 为遥测数据导出预留有限时间，避免关闭过程无限阻塞。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOtel(shutdownCtx); err != nil {
			serviceLogger.Error("shutdown OpenTelemetry", "error", err)
		}
	}()

	// 使用结构化 OTel 日志替代 Gin 默认的路由注册文本日志。
	gin.DebugPrintRouteFunc = ginotel.DebugPrintRouteFunc
	// 创建 HTTP 路由，并安装遥测、访问日志和 panic 恢复中间件。
	router := gin.New()
	router.Use(otelgin.Middleware("infra-agent"), ginotel.Logger, ginotel.Recovery)
	// 提供用于连通性检查的轻量接口。
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	// 暴露 Prometheus 指标，供采集器定期拉取。
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 将 Gin 路由挂载到可优雅关闭的标准库 HTTP Server。
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	// 缓冲通道确保服务退出时不会阻塞监听协程。
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		// 收到退出信号后停止接收新请求，并等待存量请求完成。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			serviceLogger.Error("shutdown HTTP server", "error", err)
		}
	case err := <-serverErr:
		// 除正常关闭外，记录监听服务时发生的错误。
		if !errors.Is(err, http.ErrServerClosed) {
			serviceLogger.Error("serve HTTP server", "error", err)
		}
	}
}
