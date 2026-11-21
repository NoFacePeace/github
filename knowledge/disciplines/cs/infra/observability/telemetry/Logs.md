# 日志

日志（Logs）是对系统运行中某次发生的情况的记录。例如，库存服务在查询超时时记录错误消息和当时的请求信息。

## 1. 日志记录的内容

一条日志通常包含时间戳、级别、正文和属性。例如，一条 `ERROR` 日志的正文是「库存查询超时」，属性可以记录超时阈值。如果日志包含 Trace ID 或 Span ID，还可以用它定位相关的请求或操作。

## 2. 结构化事件

[事件（Events）](Events.md)可用带事件名的结构化日志记录表示，让后端按事件名和字段筛选。[OpenTelemetry 日志数据模型](https://opentelemetry.io/docs/specs/otel/logs/data-model/)
