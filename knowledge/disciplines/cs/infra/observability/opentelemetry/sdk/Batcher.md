# Batcher

Batcher（批处理器）把多条遥测记录攒成一批，再交给 Exporter。采用 OTLP 等网络导出方式时，批量导出可以减少逐条发送产生的请求。

## 1. Span 批处理

在 SDK 中，Batch Span Processor 承担这个职责。埋点调用 `span.End()` 后，符合导出条件的 Span 会进入队列。

- 攒够一批时，处理器将这批 Span 交给 Exporter。
- 到了配置的发送时机，处理器也会交出已经攒下的 Span，即使数量还没达到批量大小。

```text
Span 结束 → 进入队列 → 按批量大小或发送时机组成一批 → Exporter
```

批处理器决定什么时候交出一批记录，Exporter 负责把这批记录导出。[OpenTelemetry 链路追踪 SDK 规范](https://opentelemetry.io/docs/specs/otel/trace/sdk/)
