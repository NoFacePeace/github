# 链路追踪

链路追踪（Traces）记录一次请求经过了哪些操作，以及这些操作之间的关系。例如，一次下单请求可以经过网关、订单服务和库存服务。

## 1. Trace 与 Span

一个 Trace 由一组关联的 [Span](Span.md) 组成，每个 Span 记录一次操作。通过 Span 之间的父子关系，可以看出一次请求经过了哪些步骤。

## 2. 查看请求耗时

假设一次下单请求花了 2.1 秒，链路显示其中库存查询 Span 耗时 2 秒，排查时就可以先查看库存查询及其下游操作。[OpenTelemetry 链路追踪](https://opentelemetry.io/docs/concepts/signals/traces/)
