# SDK

OpenTelemetry SDK 是运行在应用进程里的 OTel API 实现。埋点调用 API 记录数据，SDK 按配置处理数据。例如，业务代码每次下单成功时调用 `counter.Add(ctx, 1)`，指标 SDK 把这些增量聚合成指标数据点。

## 1. 处理不同信号

[Provider](Provider.md) 是获取对应信号的埋点工具的入口。

| 信号 | SDK 入口 | 处理示例 |
| --- | --- | --- |
| 链路追踪 | `TracerProvider` | 按采样规则记录 Span。Span 结束后，[Span Processor](Processor.md) 可以把多个 Span 攒成一批再导出 |
| 指标 | `MeterProvider` | 聚合计数或耗时等测量值，由 [Metric Reader](MetricReader.md) 收集数据点 |
| 日志 | `LoggerProvider` | 通过 Log Record Processor 处理日志记录，例如批量导出 |

## 2. 初始化与导出

SDK 通常在应用启动时初始化。使用支持自动配置的 Agent 时，也可以由 Agent 初始化。以 Go 指标 SDK 为例，创建 `MeterProvider` 时可以配置 Metric Reader。

[SDK Exporter](exporter/Exporter.md) 把 SDK 处理后的数据提供给下游。例如，OTLP Exporter 主动上报数据，Go Prometheus Exporter 作为 Metric Reader 接入 SDK，供 Prometheus 抓取指标。[OpenTelemetry 规范](https://opentelemetry.io/docs/specs/otel/)
