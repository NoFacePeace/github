# Span

Span 是链路追踪中对一次操作的记录，例如处理下单请求、调用库存服务或查询数据库。起止时间可以用来计算这次操作的耗时。

## 1. 常见字段

| 字段 | 作用 |
| --- | --- |
| 操作名 | 标明这次操作，例如「查询库存数据库」 |
| Trace ID | 标识所属的链路，同一条链路中的 Span 共享这个 ID |
| Span ID | 标识链路中的这一次操作 |
| Parent Span ID | 指向父 Span。根 Span 没有父 Span |
| 起止时间 | 记录操作何时开始和结束 |
| 属性与状态 | 记录操作细节，例如请求方法，以及操作是否出错 |

## 2. 下单链路示例

```text
Trace：一次下单请求
└── Span：订单服务处理下单
    └── Span：订单服务调用库存服务
        └── Span：库存服务处理查询
            └── Span：查询库存数据库（耗时 2 秒）
```

埋点通常在操作开始时创建 Span，在操作结束时调用 `span.End()`。在 SDK 中，已结束且符合导出条件的 Span 可以进入 Batch Span Processor 的导出队列。[OpenTelemetry 链路追踪](https://opentelemetry.io/docs/concepts/signals/traces/)
