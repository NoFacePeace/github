# 指标

指标（Metrics）是对系统运行状态或行为的数值测量。按时间记录的数据点可以用来分析请求量、错误率和延迟分布等趋势。

## 1. 常见类型

| 类型 | 表示什么 | 示例 |
| --- | --- | --- |
| 计数器（Counter） | 反复发生的事件累计了多少次 | 库存查询超时次数 |
| 仪表值（Gauge） | 某一时刻的测量值 | 当前数据库连接数 |
| 直方图（Histogram） | 多次观测值的分布 | 库存查询耗时分布 |

## 2. 指标与维度

指标可以带属性，例如下单请求计数带有 `result=success` 或 `result=error`。后端可以按结果分别聚合计数，用同一时间窗口内的失败次数和总次数算出失败率。[OpenTelemetry 指标数据模型](https://opentelemetry.io/docs/specs/otel/metrics/data-model/)
