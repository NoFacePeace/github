# SDK Exporter

SDK Exporter（应用侧导出器）在应用进程内运行，把 OpenTelemetry SDK 处理好的遥测数据提供给下游。

## 1. 数据来源与配置

SDK Exporter 通常在应用启动时接入 SDK，也可以由支持自动配置的 Agent 初始化。例如，链路追踪的 Span Processor 会把已结束的 Span 交给 SDK Exporter。

## 2. 导出方式

| 导出方式 | 谁发起请求 | 下游如何获得数据 |
| --- | --- | --- |
| OTLP Exporter | Exporter | 将遥测数据编码为 OTLP 请求，发送给 Collector 或支持 OTLP 的后端 |
| [Prometheus Exporter](Prometheus.md) | Prometheus | Exporter 将指标转换为可抓取的表示，Prometheus 定期请求 `/metrics` 读取 |
