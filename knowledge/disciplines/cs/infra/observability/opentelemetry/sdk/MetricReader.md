# Metric Reader

Metric Reader（指标读取器）负责在采集时从 OpenTelemetry SDK 收集指标数据点。比如，业务代码每次下单成功时调用 `counter.Add(ctx, 1)`，SDK 持续累加。Reader 触发采集时，可以获取「截至采集时累计下单 17 次」这样的数据点。

## 1. 谁触发采集

| 场景 | 采集时机 | 数据如何交给下游 |
| --- | --- | --- |
| OTLP 定期推送 | `PeriodicReader` 按配置的时间间隔采集 | Reader 将数据点交给 OTLP Exporter，由它主动上报 |
| Prometheus 抓取 | Prometheus 请求应用的 `/metrics` 时，Go Prometheus Exporter 触发采集 | 应用在 HTTP 响应中返回指标 |

## 2. Reader 与 Exporter

Reader 负责应用进程内的指标采集，Exporter 负责把采集结果提供给下游。Go SDK 通过 `sdkmetric.WithReader(...)` 在 `MeterProvider` 上注册 Reader。使用 OTLP 推送时，将 OTLP Exporter 交给 `PeriodicReader`。使用 Prometheus 抓取时，可以直接将 Go Prometheus Exporter 作为 Reader 注册。[OpenTelemetry 指标 SDK 规范](https://opentelemetry.io/docs/specs/otel/metrics/sdk/) · [Go Prometheus Exporter](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/prometheus)
