# Resource

Resource（资源）用一组属性描述遥测数据的来源，例如服务名、实例和部署环境。同一服务实例的指标、日志和 Span 可以关联同一组 Resource 属性。

## 1. 资源属性

以生产环境中的订单服务实例为例：

| 属性 | 示例值 | 含义 |
| --- | --- | --- |
| `service.name` | `checkout` | 服务名 |
| `service.instance.id` | `checkout-7f9d` | 服务实例标识 |
| `deployment.environment.name` | `production` | 部署环境名 |

应用可以在初始化 SDK 时配置 Resource。使用 OTLP 导出时，遥测记录按 Resource 分组，资源属性随数据一起发送。后端可以据此筛选生产环境中的订单服务数据，再按实例排查。[OpenTelemetry Resource 规范](https://opentelemetry.io/docs/specs/otel/resource/)

## 2. Resource 与 Baggage

假设订单服务 `checkout` 调用库存服务 `inventory`，并在请求中传递 `region=sg`：

| 概念 | 订单服务 | 库存服务 |
| --- | --- | --- |
| Resource | 自己产生的 Span 关联 `service.name=checkout` | 自己产生的 Span 关联 `service.name=inventory` |
| Baggage | 将 `region=sg` 写入出站请求的 `baggage` 头 | 从入站请求读取同一个 `region=sg`，也可以继续向下游传递 |

Resource 属性随遥测记录导出。要在后端按 Baggage 中的 `region` 筛选链路，可以让埋点将这个值写入 Span 属性。[OpenTelemetry Baggage](https://opentelemetry.io/docs/concepts/signals/baggage/)
