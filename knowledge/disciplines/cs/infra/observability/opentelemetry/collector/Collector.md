# OpenTelemetry Collector

OpenTelemetry Collector 是独立运行的遥测采集程序。应用可以把遥测数据发给 Collector，由它集中处理并转发。后端支持应用所用的 OTLP 传输方式和信号时，应用也可以直接上报。

## 1. Collector 的流水线

| 组件 | 职责 | 示例 |
| --- | --- | --- |
| Receiver（接收器） | 将上游数据引入流水线 | OTLP Receiver 接收应用的推送，Prometheus Receiver 定期抓取指标 |
| Processor（处理器） | 对进入流水线的数据执行配置的处理 | 过滤健康检查 Span、删除属性、批处理 |
| [Exporter（导出器）](exporter/Exporter.md) | 把流水线数据交给下游 | OTLP Exporter 主动上报，Prometheus Exporter 暴露指标供抓取 |

Collector 是整个程序，Receiver 是程序内的数据入口。一个 Collector 可以配置多条流水线，每条流水线包含接收器、可选的处理器和导出器。[Collector 架构](https://opentelemetry.io/docs/collector/architecture/)

## 2. 直连与中转

```text
应用 OTLP Exporter ──OTLP──→ 支持 OTLP 的后端

应用 OTLP Exporter ──OTLP──→ Collector
                             Receiver → Processor → Exporter ──→ 后端
```

当多个应用需要共用处理规则时，可以把规则放在 Collector。例如，让各应用发送到同一个 Collector，由它统一过滤属性，把指标发给 Prometheus，把链路发给 Tempo。后端地址变化时，可以在 Collector 更新对应的导出配置。
