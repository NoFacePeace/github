package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/signal"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/gin-gonic/gin"
)

var tracer = otel.Tracer("otel")

func main() {
	ctx := signal.SetupSignalHandler()

	tp, err := initTracer()
	if err != nil {
		slog.Error("init tracer error", "error", err)
		os.Exit(1)
	}
	r := gin.Default()
	r.Use(otelgin.Middleware("gin-middleware"))
	r.GET("/ping", func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "ping")
		defer span.End()
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
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tp.Shutdown(ctx); err != nil {
		slog.Error("tracer provider shutdown error", "error", err)
	}
}

func initTracer() (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("stdouttrace new error: [%w]", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func initMeter() {
	// exporter, err := prometheus.New()
	// tp := metrics.New
	// otel.SetLogger()
	// otelslog.NewLogger()
	// otelslog.NewLogger(name)
	// slog.SetDefault()
}
