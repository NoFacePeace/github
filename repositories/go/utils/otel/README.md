# OpenTelemetry 初始化工具

`otel.Setup` 用于在应用的 `main` 函数中初始化全局的 Trace、Metric 和 Log
Provider。每个进程只能调用一次；业务包仅通过 `otel.Tracer`、`otel.Meter` 等
全局 API 获取对应对象，不应重复初始化 Provider。`With...` Option 只收集配置
和组件引用，不会在执行时修改任何全局 Provider。

## 使用方式

```go
ctx := context.Background()
shutdown, err := utilotel.Setup(ctx,
	utilotel.WithServiceName("payment-api"),
	utilotel.WithOtlptracehttp(
		otlptracehttp.WithEndpoint("otel-collector:4318"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Bearer <token>",
		}),
	),
	utilotel.WithMetricPrometheus(),
)
if err != nil {
	return err
}
defer func() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown OpenTelemetry failed", "error", err)
	}
}()
```

`WithOtlptracehttp` 和 `WithMetricPrometheus` 是保留的快捷 Option。前者无参数
时与旧版一致，使用默认 endpoint 并允许明文 HTTP；生产环境应显式传入 TLS 选项。
后者可配合 `promhttp.Handler()` 暴露 `/metrics`。

未设置任意 Trace exporter 或 Metric Reader Option 时，对应信号会输出到标准输出，
适合本地调试。OTLP/gRPC 或其他供应商 exporter 可通过 `WithTraceExporter` 传入；
Metric 也可通过 `WithMetricReader` 传入 OTLP PeriodicReader 或其他 SDK Reader。

所有信号共享同一份 Resource，因此 `ServiceName` 会同时写入 Trace、Metric 和
Log。关闭函数是幂等的，会以反向顺序关闭 Provider，并汇总全部关闭错误。成功
调用 `Setup` 后，Provider 会接管传入 exporter、reader 与 processor 的关闭。
