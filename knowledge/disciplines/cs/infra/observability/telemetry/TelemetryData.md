# 遥测数据

遥测数据（Telemetry Data）是系统运行时采集的关于其状态和行为的数据，例如请求耗时、错误记录和调用链路。

## 1. 遥测数据与信号

[信号](Signal.md)指遥测数据的类别。例如，「日志」是一种信号，「10:03 库存查询超时」是一条具体的遥测记录。

## 2. 关联日志和链路

一个 Trace 由一组关联的 Span 组成，每个 Span 表示一次操作。如果超时日志带有 Trace ID，就可以用它找到同一次请求的链路，查看库存查询前后的调用和耗时。[OpenTelemetry 可观测性入门](https://opentelemetry.io/docs/concepts/observability-primer/)
