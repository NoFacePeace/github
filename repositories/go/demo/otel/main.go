package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	utilginotel "github.com/NoFacePeace/github/repositories/go/utils/gin/otel"
	utilotel "github.com/NoFacePeace/github/repositories/go/utils/otel"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
)

var (
	logger = otelslog.NewLogger("demo", otelslog.WithSource(true))
	tracer = otel.Tracer("roll")
)

func main() {
	ctx := signal.SetupSignalHandler()
	otelShutdown, err := utilotel.Setup(ctx, utilotel.WithServiceName("main"), utilotel.WithMetricPrometheus(), utilotel.WithOtlptracehttp())
	if err != nil {
		slog.Error("otel set up error", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(ctx); err != nil {
			slog.Error("otel shutdown error", "error", err)
			return
		}
		slog.Info("otel shudown success")
	}()
	// gin.SetMode(gin.ReleaseMode)
	// r := gin.Default()
	r := gin.New()
	r.Use(otelgin.Middleware("gin"))
	r.Use(utilginotel.Logger, utilginotel.Recovery)
	r.Use(utilginotel.InjectPropagatorToResponseHeader)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/ping", func(c *gin.Context) {
		logger.InfoContext(c.Request.Context(), "test")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	srv := &http.Server{
		Handler: r,
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("http server shutdown error", "error", err)
			return
		}
		slog.Info("http server shutdown success")
	}()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("http server listen and server error", "error", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
}
