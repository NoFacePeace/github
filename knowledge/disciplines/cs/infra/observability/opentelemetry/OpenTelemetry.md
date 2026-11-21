# OpenTelemetry

OpenTelemetry（OTel）是一套用于生成、采集、处理和导出遥测数据的标准与工具。规范涵盖 API、SDK 行为、数据模型、语义约定和传输协议。后端负责存储和查询。

## 1. 从埋点到导出

| 概念 | 职责 | 下单计数示例 |
| --- | --- | --- |
| Instrumentation（埋点或插桩） | 决定在什么业务事件上记录什么 | 在下单成功后调用 `counter.Add(ctx, 1)` |
| API | 给埋点代码提供接口 | 提供创建计数器和记录增量的接口 |
| [SDK](sdk/SDK.md) | 实现 API，处理应用进程内的遥测数据 | 将多次 `+1` 聚合为指标数据点 |
| 数据模型 | 定义信号有哪些字段以及如何解释这些字段 | 用 [Resource](Resource.md) 表示指标来源，用数据点表示时间戳、值等信息 |
| 语义约定 | 统一常见字段的名称和含义 | 用 `service.name` 表示服务名 |
| SDK Exporter（应用侧导出器） | 通过所配置的方式把 SDK 数据提供给下游 | 主动上报计数，或在抓取时提供计数 |

业务代码可以手动调用 OTel API，也可以接入为 HTTP 框架、数据库客户端等提供的自动埋点。

## 2. 传输与采集

- [OTLP](OTLP.md) 定义遥测数据如何编码和主动上报。OTLP Exporter 实现这一发送方式。
- [Collector](collector/Collector.md) 可以作为应用的接收端，集中处理和转发数据。支持相应信号和传输方式的后端也可以直接接收。

## 3. 上下文传播

[上下文传播](propagation/Propagation.md)让下游服务接续同一条调用链，并读取上游传来的自定义请求上下文。

## 4. Bridge

[Bridge（桥接）](Bridge.md)把已有的埋点接口或遥测数据接入 OpenTelemetry。
