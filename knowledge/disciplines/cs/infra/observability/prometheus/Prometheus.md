# Prometheus

Prometheus 是指标后端。它把接收到的指标存成时间序列，并提供 PromQL 查询。OpenTelemetry 指标可以通过 Prometheus 抓取或 OTLP 上报接入。

## 1. Prometheus 抓取

Prometheus 定期向配置的目标发送 HTTP GET 请求，读取 `/metrics` 上以 Prometheus 或 OpenMetrics 格式暴露的指标。抓取目标可以是应用，也可以是 Collector。

## 2. Prometheus 接收 OTLP

Prometheus 启动时启用 `--web.enable-otlp-receiver` 后，应用里的 OTLP/HTTP Exporter 可以把指标直接发送到 Prometheus 的 `/api/v1/otlp/v1/metrics`。Prometheus 将收到的指标转换为自己的时间序列。[Prometheus 接收 OpenTelemetry 指标](https://prometheus.io/docs/guides/opentelemetry/)

## 3. 指标模型与格式

[Prometheus 指标](metrics/Metrics.md)介绍 Counter、Gauge、Histogram 和 Summary 四种常见类型及其查询方式。

OpenTelemetry 和 Prometheus 都有计数器和直方图，但各自定义了数据模型。假设两个采集周期分别新增 10 单和 7 单。OTel 可以将它们表示成增量值 `10、7`，也可以用累计值 `10、17`。Prometheus 计数器通常以累计值呈现。[OpenTelemetry 指标数据模型](https://opentelemetry.io/docs/specs/otel/metrics/data-model/)
