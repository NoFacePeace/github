# Baggage

Baggage 是随请求上下文传播的自定义键值对。例如，网关可以把 `region=sg` 放入 Baggage，通过 W3C `baggage` 请求头传给订单服务。订单服务调用库存服务时，可以继续转发这个值。

## 1. 将 Baggage 写入遥测数据

Baggage 中的值可供下游埋点读取。要在后端按 `region` 筛选链路，可以通过埋点或配置把这个值写入 Span 属性。[OpenTelemetry Baggage](https://opentelemetry.io/docs/concepts/signals/baggage/)
