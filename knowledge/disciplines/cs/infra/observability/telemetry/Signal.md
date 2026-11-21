# 信号

信号（Signal）是遥测数据的类别。每种信号以不同的方式描述系统的运行情况。同一次库存查询超时可以同时产生指标、日志和链路追踪数据。

## 1. 常见信号

| 信号 | 数据示例 | 能回答的问题 |
| --- | --- | --- |
| [指标（Metrics）](Metrics.md) | 超时计数从 `17` 增至 `18` | 超时有多少，何时开始变多？ |
| [日志（Logs）](Logs.md) | 一条记录包含超时时间、错误消息和 Trace ID | 这次超时发生时有哪些细节？ |
| [链路追踪（Traces）](Traces.md) | 下单请求的链路中，库存查询 Span 耗时 2 秒 | 请求经过哪些服务，时间花在哪一步？ |
| [性能剖析（Profiles）](Profiles.md) | 采样显示订单服务的 JSON 编码函数占用了较多 CPU | CPU 或内存消耗集中在哪些函数？ |

OpenTelemetry 的信号文档也列出了 Baggage。它携带随请求传播的自定义上下文，例如 `region=sg`。[OpenTelemetry 信号文档](https://opentelemetry.io/docs/concepts/signals/)
