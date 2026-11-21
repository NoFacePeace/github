# 上下文传播

上下文传播（Propagation）是把请求上下文从调用方传给被调用方的过程。以订单服务调用库存服务为例，订单服务把上下文写入出站 HTTP 请求头，库存服务从入站请求头恢复上下文。

## 1. 传播的内容

| 上下文 | HTTP 请求头 | 下游用途 |
| --- | --- | --- |
| Trace Context（链路上下文） | `traceparent`，以及可选的 `tracestate` | 使用上游的链路和 Span 信息，把库存查询接到同一条 Trace 上 |
| [Baggage](Baggage.md) | `baggage`，例如 `baggage: region=sg` | 读取上游传来的自定义键值对 |

## 2. 传播器

[Propagator（传播器）](Propagator.md)负责在应用上下文和 HTTP 请求头等载体之间读写可传播的信息。[OpenTelemetry 上下文传播](https://opentelemetry.io/docs/concepts/context-propagation/)
