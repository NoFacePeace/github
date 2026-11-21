# 事件

在 OpenTelemetry 的日志模型中，事件（Event）是带事件名的日志记录。事件名标识发生了哪类事情，属性描述这一次发生时的细节。

## 1. 结构化记录示例

库存服务可以记录名为 `inventory.query.timeout` 的事件，在属性里填写超时阈值和重试次数。查询时可以先按事件名筛选，再按这些属性细分。[OpenTelemetry 日志数据模型](https://opentelemetry.io/docs/specs/otel/logs/data-model/)
