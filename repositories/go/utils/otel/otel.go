package otel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/bridges/otellogr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	resource   *resource.Resource
	propagator propagation.TextMapPropagator
	logger     log.LoggerProvider
	tracer     trace.TracerProvider
	meter      metric.MeterProvider
}

type Option func(ctx context.Context, cfg *config) (func(context.Context) error, error)

func Setup(ctx context.Context, options ...Option) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var err error
		for _, f := range shutdownFuncs {
			err = errors.Join(f(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	var err error
	handleErr := func(e error) {
		err = errors.Join(e, shutdown(ctx))
	}
	cfg := &config{}
	cfg.resource = resource.Default()
	for _, opt := range options {
		sd, err := opt(ctx, cfg)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, sd)
	}
	if cfg.propagator == nil {
		setPropagator(cfg)
	}
	if cfg.logger == nil {
		sd, err := setLogger(cfg)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, sd)
	}
	if cfg.tracer == nil {
		sd, err := setTracer(cfg)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, sd)
	}
	if cfg.meter == nil {
		sd, err := setMeter(cfg)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, sd)
	}
	return shutdown, err
}

func setPropagator(cfg *config) {
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)
	cfg.propagator = prop
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func setLogger(cfg *config) (func(ctx context.Context) error, error) {
	exporter, err := stdoutlog.New()
	if err != nil {
		return nil, fmt.Errorf("stdoutlog new error: [%w]", err)
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(cfg.resource),
	)
	setInternalLogger()
	global.SetLoggerProvider(provider)
	cfg.logger = provider
	return provider.Shutdown, nil
}

// 设置 otel 内部 logger, 不能设置属性，为触发死锁 sync.once
func setInternalLogger() {
	provider := global.GetLoggerProvider()
	if provider != nil {
		return
	}
	logger := logr.New(otellogr.NewLogSink("otel", otellogr.WithLoggerProvider(provider)))
	otel.SetLogger(logger)
}

func setTracer(cfg *config) (func(ctx context.Context) error, error) {
	exporter, err := stdouttrace.New()
	if err != nil {
		return nil, fmt.Errorf("stdouttrace new error: [%w]", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(cfg.resource),
	)

	otel.SetTracerProvider(provider)
	cfg.tracer = provider
	return provider.Shutdown, nil
}

func setMeter(cfg *config) (func(ctx context.Context) error, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("stdoutmetric new error: [%w]", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(3*time.Second))),
		sdkmetric.WithResource(cfg.resource),
	)
	otel.SetMeterProvider(provider)
	cfg.meter = provider
	return provider.Shutdown, nil
}

func WithServiceName(name string) func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
	return func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
		res := cfg.resource

		res1, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(name)))
		if err != nil {
			return nil, fmt.Errorf("resource new errror: [%w]", err)
		}
		res2, err := resource.Merge(res, res1)
		if err != nil {
			return nil, fmt.Errorf("resource merge error: [%w]", err)
		}
		*res = *res2
		return func(ctx context.Context) error { return nil }, nil
	}
}

func WithMetricPrometheus() func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
	return func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
		exporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("prometheus new error: [%w]", err)
		}
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
		otel.SetMeterProvider(provider)
		cfg.meter = provider
		return provider.Shutdown, nil
	}
}

func WithOtlptracehttp() func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
	return func(ctx context.Context, cfg *config) (func(context.Context) error, error) {
		exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("otlp trace http new error: [%w]", err)
		}
		provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
		otel.SetTracerProvider(provider)
		cfg.tracer = provider
		return provider.Shutdown, nil
	}
}
