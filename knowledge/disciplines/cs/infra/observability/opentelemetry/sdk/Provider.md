# Provider

在 OpenTelemetry 中，Provider 提供用于记录链路、指标或日志的对象。应用初始化 SDK 时可以配置对应的 Provider，埋点代码再从它获取所需的工具。

## 1. 常见 Provider

| Provider | 提供的对象 | 用途 |
| --- | --- | --- |
| `TracerProvider` | `Tracer` | 开始一个 Span |
| `MeterProvider` | `Meter` | 创建计数器、直方图等指标工具 |
| `LoggerProvider` | `Logger` | 提交日志记录 |

## 2. 下单计数示例

Go 应用启动时创建 `MeterProvider`，并为它配置 Metric Reader。埋点代码从 Provider 获取 `Meter`，再通过 `Meter` 创建下单计数器。之后每次下单成功时，使用这个计数器记录一次增量。

```text
MeterProvider → Meter → Counter → counter.Add(ctx, 1)
```
