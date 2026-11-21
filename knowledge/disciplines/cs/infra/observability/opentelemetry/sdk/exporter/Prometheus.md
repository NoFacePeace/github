# Prometheus Exporter

Go 的 Prometheus Exporter 将应用进程内的 OTel SDK 指标提供给 Prometheus 抓取。比如，下单计数可以呈现为 `orders_total 17`。

## 1. 接入 Go SDK

应用用 `go.opentelemetry.io/otel/exporters/prometheus` 包的 `New()` 创建 Exporter，通过 `sdkmetric.WithReader(...)` 将它作为 Metric Reader 接到 `MeterProvider`，再用 `promhttp.Handler()` 提供 `/metrics`。当 Prometheus 抓取应用的 `/metrics` 时，Exporter 读取 SDK 聚合的指标，并将其转换成可抓取的表示。[Go Prometheus Exporter 文档](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/prometheus)
