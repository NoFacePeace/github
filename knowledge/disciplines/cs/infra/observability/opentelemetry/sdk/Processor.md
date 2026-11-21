# Processor

Processor（处理器）在 OpenTelemetry SDK 中处理 Span 或日志记录。例如，它可以把多条记录攒成一批，再交给 Exporter。

## 1. Span Processor

Span Processor 可以在 Span 开始时收到 `OnStart` 回调，在 Span 结束时收到 `OnEnd` 回调。Batch Span Processor 通过[批处理](Batcher.md)将符合导出条件的 Span 成批交给 Exporter。[OpenTelemetry 链路追踪 SDK 规范](https://opentelemetry.io/docs/specs/otel/trace/sdk/)

## 2. Log Record Processor

日志通过 OTel Logger 提交时，SDK 调用 Log Record Processor 的 `OnEmit`。处理器可以补充日志属性，或者把日志加入队列，稍后批量导出。[OpenTelemetry 日志 SDK 规范](https://opentelemetry.io/docs/specs/otel/logs/sdk/)
