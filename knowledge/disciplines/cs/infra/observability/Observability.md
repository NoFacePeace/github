# 可观测性

可观测性（Observability）是通过系统对外产生的信号推断其内部状态的能力。监控常用预先设定的指标和告警发现异常。排查异常时，可以结合指标、日志和链路追踪判断影响范围、定位问题发生的位置，并验证可能的原因。

## 1. 遥测数据与信号

[遥测数据](telemetry/TelemetryData.md)是系统运行时采集的数据。信号（Signal）指遥测数据的类别，常见的有指标、日志和链路追踪。

## 2. 采集与查询

[OpenTelemetry](opentelemetry/OpenTelemetry.md) 提供埋点、采集、处理和导出遥测数据的标准与工具。查询语言和 API 由后端定义。例如，[Prometheus](prometheus/Prometheus.md) 用 PromQL 计算过去一小时的下单速率，Tempo 用 TraceQL 查找库存查询超时的链路。

Prometheus 定期读取 `/metrics` 是采集，读到的是目标当前暴露的指标。通过 PromQL 查看过去一小时的速率是查询，读的是 Prometheus 已保存的时间序列。

[VictoriaMetrics](victoriametrics/VictoriaMetrics.md) 可以集中存储多个采集端写入的指标，并通过 MetricsQL 查询。
