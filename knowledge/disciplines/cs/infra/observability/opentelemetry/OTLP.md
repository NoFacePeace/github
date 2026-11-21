# OTLP

OTLP（OpenTelemetry Protocol）定义遥测数据在导出请求中的表示，以及发送端和接收端交换请求的规则。同一份 SDK 指标可以通过 OTLP Exporter 编码并发送，也可以通过 Prometheus Exporter 供抓取。

## 1. 导出请求里的数据

链路、指标和日志各有自己的导出请求。以链路为例，一个 `ExportTraceServiceRequest` 可以包含多个资源和埋点范围，每组下面有多条 Span。

| 层级 | 内容 | 示例 |
| --- | --- | --- |
| Resource（资源） | 遥测数据来源的属性 | `service.name=checkout` |
| Instrumentation Scope（埋点范围） | 生成记录的埋点库或组件的信息 | HTTP 埋点库的名称和版本 |
| 遥测记录 | 具体的 Span、指标数据或日志记录 | 库存查询 Span、下单计数、查询超时日志 |

Span 可以包含 Trace ID、Span ID、父 Span ID、起止时间、属性和状态。指标包含名称、单位、类型和数据点。日志记录可以包含时间戳、级别、正文、属性，以及用于关联链路的 Trace ID 和 Span ID。

## 2. 批次与 Protobuf package

- **批次**对应一次导出请求。一次请求可以带多个 Trace 的 Span，同一个 Trace 的 Span 也可以通过多次请求上报。
- **Protobuf `package`** 是定义消息和服务的命名空间。例如，`opentelemetry.proto.collector.trace.v1` 定义链路导出服务和请求，`opentelemetry.proto.trace.v1` 定义 Span 等数据结构。[导出服务定义](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/collector/trace/v1/trace_service.proto) · [Span 结构定义](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto)

## 3. 传输方式

OTLP/gRPC 和 OTLP/HTTP 都由发送端发起导出请求，接收端返回结果。默认端口分别是 `4317` 和 `4318`。OTLP/HTTP 的常见默认路径是 `/v1/traces`、`/v1/metrics` 和 `/v1/logs`。[OTLP 协议规范](https://opentelemetry.io/docs/specs/otlp/)
