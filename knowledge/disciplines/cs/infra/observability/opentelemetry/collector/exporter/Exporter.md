# Exporter

Exporter 是 OpenTelemetry Collector 流水线的出口。它接收从 Receiver 进入、经过可选 Processor 处理的数据，再把数据提供给下游。

## 1. 配置位置

在 Collector 配置文件的 `exporters` 下定义导出器，再在对应信号的 `service.pipelines.<signal>.exporters` 列表中引用它。例如，可以让 traces 流水线使用 OTLP Exporter 把 Span 发送到链路后端。[Collector 架构](https://opentelemetry.io/docs/collector/architecture/)

## 2. 导出方式

- **OTLP Exporter** 主动把 Collector 流水线中的遥测数据发送到下游，例如把从应用收到的 Span 转发给后端。
- [Prometheus Exporter](Prometheus.md) 在 Collector 上暴露指标端点，供 Prometheus 定期抓取。
