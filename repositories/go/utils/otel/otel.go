// Package otel 负责在程序入口初始化 OpenTelemetry。
//
// Setup 会创建并安装全局的 Trace、Metric、Log Provider，因此每个进程只能在
// main 中调用一次。库代码应从全局 Provider 获取 Tracer 或 Meter，而不应调用
// Setup 覆盖应用已经安装的 Provider。
package otel

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/bridges/otellogr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Option 仅描述 OpenTelemetry 的初始化配置，不能修改 OpenTelemetry 全局状态。
// Setup 会在全部 Option 应用成功后，统一创建并安装 Provider。
type Option func(*config) error

type config struct {
	serviceName     string
	resources       []*resource.Resource
	traceExporters  []sdktrace.SpanExporter
	traceFactories  []traceExporterFactory
	metricReaders   []sdkmetric.Reader
	metricFactories []metricReaderFactory
	logProcessors   []sdklog.Processor
	logFactories    []logProcessorFactory
}

// traceExporterFactory 在 Setup 阶段创建 Trace exporter，避免 Option 应用时产生
// 网络连接或其他初始化副作用。
type traceExporterFactory func(context.Context) (sdktrace.SpanExporter, error)

// metricReaderFactory 在 Setup 阶段创建 Metric Reader。
type metricReaderFactory func(context.Context) (sdkmetric.Reader, error)

// logProcessorFactory 在 Setup 阶段创建 Log Processor。
type logProcessorFactory func(context.Context) (sdklog.Processor, error)

// WithServiceName 设置 service.name 资源属性。该属性会同时应用于 Trace、Metric
// 与 Log；若同时传入携带 service.name 的 Resource，以此 Option 的值为准。
func WithServiceName(name string) Option {
	return func(cfg *config) error {
		if name == "" {
			return errors.New("service name cannot be empty")
		}
		cfg.serviceName = name
		return nil
	}
}

// WithResource 合并额外的 Resource，例如部署环境、服务版本或云资源属性。
// Resource 会和 SDK 默认 Resource 合并，调用方传入的 Resource 不会被修改。
func WithResource(res *resource.Resource) Option {
	return func(cfg *config) error {
		if res == nil {
			return errors.New("resource cannot be nil")
		}
		cfg.resources = append(cfg.resources, res)
		return nil
	}
}

// WithTraceExporter 添加一个 Trace exporter。未添加时，Trace 会输出到标准输出，
// 便于本地调试。可以多次调用以同时导出到多个后端。
//
// Setup 成功后会接管 exporter 的生命周期，并在 shutdown 时关闭它。
func WithTraceExporter(exporter sdktrace.SpanExporter) Option {
	return func(cfg *config) error {
		if exporter == nil {
			return errors.New("trace exporter cannot be nil")
		}
		cfg.traceExporters = append(cfg.traceExporters, exporter)
		return nil
	}
}

// WithOtlptracehttp 添加 OTLP/HTTP Trace exporter。没有传入选项时，行为与旧版
// 一致：使用环境变量或默认 endpoint，并允许明文 HTTP。生产环境应显式传入 TLS
// 相关选项，例如 otlptracehttp.WithTLSClientConfig。
//
// 该 Option 仅保存 exporter 工厂；实际 exporter 在 Setup(ctx, ...) 中创建。
func WithOtlptracehttp(options ...otlptracehttp.Option) Option {
	optionsCopy := append([]otlptracehttp.Option(nil), options...)
	if len(optionsCopy) == 0 {
		optionsCopy = []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	}
	return func(cfg *config) error {
		cfg.traceFactories = append(cfg.traceFactories, func(ctx context.Context) (sdktrace.SpanExporter, error) {
			exporter, err := otlptracehttp.New(ctx, optionsCopy...)
			if err != nil {
				return nil, fmt.Errorf("create OTLP HTTP trace exporter: %w", err)
			}
			return exporter, nil
		})
		return nil
	}
}

// WithMetricReader 添加一个 Metric Reader，例如 Prometheus exporter 或使用
// PeriodicReader 包装的 OTLP Metric exporter。未添加时，Metric 输出到标准输出。
// 可以多次调用以向多个后端提供 Metric。
//
// Setup 成功后会接管 reader 的生命周期，并在 shutdown 时关闭它。
func WithMetricReader(reader sdkmetric.Reader) Option {
	return func(cfg *config) error {
		if reader == nil {
			return errors.New("metric reader cannot be nil")
		}
		cfg.metricReaders = append(cfg.metricReaders, reader)
		return nil
	}
}

// WithMetricPrometheus 添加 Prometheus Metric Reader。调用方可通过
// promhttp.Handler 暴露 /metrics。该 Option 不会在应用时修改全局 MeterProvider。
func WithMetricPrometheus() Option {
	return func(cfg *config) error {
		cfg.metricFactories = append(cfg.metricFactories, func(context.Context) (sdkmetric.Reader, error) {
			reader, err := prometheus.New()
			if err != nil {
				return nil, fmt.Errorf("create Prometheus metric reader: %w", err)
			}
			return reader, nil
		})
		return nil
	}
}

// WithLogProcessor 添加一个 OpenTelemetry Log Processor。未添加时，Log 会通过
// BatchProcessor 输出到标准输出。可以多次调用以组合多个 Processor。
//
// Setup 成功后会接管 processor 的生命周期，并在 shutdown 时关闭它。
func WithLogProcessor(processor sdklog.Processor) Option {
	return func(cfg *config) error {
		if processor == nil {
			return errors.New("log processor cannot be nil")
		}
		cfg.logProcessors = append(cfg.logProcessors, processor)
		return nil
	}
}

// WithOtlploghttp 添加 OTLP/HTTP Log exporter。没有传入选项时，使用环境变量或
// 默认 endpoint，并允许明文 HTTP。调用方可使用 otlploghttp.WithEndpointURL 配置
// Loki 的 OTLP endpoint。
//
// 该 Option 仅保存 Processor 工厂；实际 exporter 在 Setup(ctx, ...) 中创建。
func WithOtlploghttp(options ...otlploghttp.Option) Option {
	optionsCopy := append([]otlploghttp.Option(nil), options...)
	if len(optionsCopy) == 0 {
		optionsCopy = []otlploghttp.Option{otlploghttp.WithInsecure()}
	}
	return func(cfg *config) error {
		cfg.logFactories = append(cfg.logFactories, func(ctx context.Context) (sdklog.Processor, error) {
			exporter, err := otlploghttp.New(ctx, optionsCopy...)
			if err != nil {
				return nil, fmt.Errorf("create OTLP HTTP log exporter: %w", err)
			}
			return sdklog.NewBatchProcessor(exporter), nil
		})
		return nil
	}
}

// Setup 构建并安装 OpenTelemetry 全局 Provider，返回用于刷出缓存并关闭 Provider
// 的函数。调用方应在 main 中只调用一次，并使用带超时的 context 调用 shutdown。
// 所有 Provider 会共享同一份 Resource；在全部创建成功之前不会修改全局状态。
func Setup(ctx context.Context, options ...Option) (func(context.Context) error, error) {
	cfg := config{}
	for _, opt := range options {
		if opt == nil {
			return nil, errors.New("OpenTelemetry option cannot be nil")
		}
		if err := opt(&cfg); err != nil {
			return nil, fmt.Errorf("apply OpenTelemetry option: %w", err)
		}
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, err
	}
	var shutdownFuncs []func(context.Context) error
	fail := func(err error) (func(context.Context) error, error) {
		return nil, errors.Join(err, shutdownAll(context.Background(), shutdownFuncs))
	}

	traceProvider, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return fail(err)
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)
	meterProvider, err := newMeterProvider(ctx, cfg, res)
	if err != nil {
		return fail(err)
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	loggerProvider, err := newLoggerProvider(ctx, cfg, res)
	if err != nil {
		return fail(err)
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)

	// 仅在所有组件创建成功后才写入全局状态，避免留下半初始化状态。
	otel.SetTextMapPropagator(newPropagator())
	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(meterProvider)
	global.SetLoggerProvider(loggerProvider)
	otel.SetLogger(logr.New(otellogr.NewLogSink("otel", otellogr.WithLoggerProvider(loggerProvider))))
	return newShutdown(shutdownFuncs), nil
}

// newResource 返回本次初始化专用的 Resource，不会修改 resource.Default()。
func newResource(cfg config) (*resource.Resource, error) {
	res := resource.Default()
	var err error
	for _, extra := range cfg.resources {
		res, err = resource.Merge(res, extra)
		if err != nil {
			return nil, fmt.Errorf("merge resource: %w", err)
		}
	}
	if cfg.serviceName == "" {
		return res, nil
	}
	return resource.Merge(res, resource.NewWithAttributes("", semconv.ServiceName(cfg.serviceName)))
}

// newPropagator 同时启用 W3C Trace Context 与 Baggage 跨服务传播。
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

// newTracerProvider 使用所有配置的 exporter 创建 Provider。
func newTracerProvider(ctx context.Context, cfg config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporters := append([]sdktrace.SpanExporter(nil), cfg.traceExporters...)
	var created []sdktrace.SpanExporter
	for _, factory := range cfg.traceFactories {
		exporter, err := factory(ctx)
		if err != nil {
			return nil, errors.Join(err, shutdownTraceExporters(ctx, created))
		}
		created = append(created, exporter)
		exporters = append(exporters, exporter)
	}
	if len(exporters) == 0 {
		exporter, err := stdouttrace.New()
		if err != nil {
			return nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		exporters = []sdktrace.SpanExporter{exporter}
	}
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	for _, exporter := range exporters {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}
	return sdktrace.NewTracerProvider(opts...), nil
}

// newMeterProvider 使用所有配置的 Reader 创建 Provider。
func newMeterProvider(ctx context.Context, cfg config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	readers := append([]sdkmetric.Reader(nil), cfg.metricReaders...)
	var created []sdkmetric.Reader
	for _, factory := range cfg.metricFactories {
		reader, err := factory(ctx)
		if err != nil {
			return nil, errors.Join(err, shutdownMetricReaders(ctx, created))
		}
		created = append(created, reader)
		readers = append(readers, reader)
	}
	if len(readers) == 0 {
		exporter, err := stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("create stdout metric exporter: %w", err)
		}
		readers = []sdkmetric.Reader{sdkmetric.NewPeriodicReader(exporter)}
	}
	opts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	for _, reader := range readers {
		opts = append(opts, sdkmetric.WithReader(reader))
	}
	return sdkmetric.NewMeterProvider(opts...), nil
}

// newLoggerProvider 使用所有配置的 Processor 创建 Provider。
func newLoggerProvider(ctx context.Context, cfg config, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	processors := append([]sdklog.Processor(nil), cfg.logProcessors...)
	var created []sdklog.Processor
	for _, factory := range cfg.logFactories {
		processor, err := factory(ctx)
		if err != nil {
			return nil, errors.Join(err, shutdownLogProcessors(ctx, created))
		}
		created = append(created, processor)
		processors = append(processors, processor)
	}
	if len(processors) == 0 {
		exporter, err := stdoutlog.New()
		if err != nil {
			return nil, fmt.Errorf("create stdout log exporter: %w", err)
		}
		processors = []sdklog.Processor{sdklog.NewBatchProcessor(exporter)}
	}
	opts := []sdklog.LoggerProviderOption{sdklog.WithResource(res)}
	for _, processor := range processors {
		opts = append(opts, sdklog.WithProcessor(processor))
	}
	return sdklog.NewLoggerProvider(opts...), nil
}

// newShutdown 返回幂等的关闭函数。Provider 按创建的反序关闭，且保留所有错误。
func newShutdown(shutdownFuncs []func(context.Context) error) func(context.Context) error {
	var once sync.Once
	var shutdownErr error
	return func(ctx context.Context) error {
		once.Do(func() { shutdownErr = shutdownAll(ctx, shutdownFuncs) })
		return shutdownErr
	}
}

// shutdownAll 反向关闭所有 Provider，用于初始化失败时清理已经创建的组件。
func shutdownAll(ctx context.Context, shutdownFuncs []func(context.Context) error) error {
	var err error
	for _, shutdownFunc := range slices.Backward(shutdownFuncs) {
		err = errors.Join(err, shutdownFunc(ctx))
	}
	return err
}

// shutdownTraceExporters 关闭尚未交给 Provider 管理的 exporter。
func shutdownTraceExporters(ctx context.Context, exporters []sdktrace.SpanExporter) error {
	var err error
	for _, exporter := range slices.Backward(exporters) {
		err = errors.Join(err, exporter.Shutdown(ctx))
	}
	return err
}

// shutdownMetricReaders 关闭尚未交给 Provider 管理的 Reader。
func shutdownMetricReaders(ctx context.Context, readers []sdkmetric.Reader) error {
	var err error
	for _, reader := range slices.Backward(readers) {
		err = errors.Join(err, reader.Shutdown(ctx))
	}
	return err
}

// shutdownLogProcessors 关闭尚未交给 Provider 管理的 Log Processor。
func shutdownLogProcessors(ctx context.Context, processors []sdklog.Processor) error {
	var err error
	for _, processor := range slices.Backward(processors) {
		err = errors.Join(err, processor.Shutdown(ctx))
	}
	return err
}
