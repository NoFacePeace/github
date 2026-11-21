# Propagator

Propagator（传播器）是实现上下文传播的组件。它决定读写哪些字段，以及这些字段在载体中的格式。HTTP 请求头就是一种载体。

## 1. 传播器的类型

以 OpenTelemetry Go 为例，`propagation.TraceContext{}` 负责 `traceparent` 和 `tracestate`，`propagation.Baggage{}` 负责 `baggage`。可以用 `propagation.NewCompositeTextMapPropagator` 把两者组合起来，让一次 HTTP 调用同时传播链路信息和自定义键值对。[OpenTelemetry Go propagation 包](https://pkg.go.dev/go.opentelemetry.io/otel/propagation)

## 2. 传播操作

一次 HTTP 调用中，发送方先用 [Inject](Inject.md) 将上下文写入载体，接收方再用 [Extract](Extract.md) 从载体恢复上下文。
