# Bridge

Bridge（桥接）将已有的埋点接口或产生的遥测数据接入 OpenTelemetry，使它们可以和原生 OTel 埋点共用导出路径。

## 1. 常见桥接示例

- **OpenTracing Bridge** 将 OpenTracing API 调用转给 OpenTelemetry。例如，旧代码调用 `StartSpan()` 时，Go 的 `bridge/opentracing` 将它适配为 OTel 的 `Start()` 调用。[OpenTracing Bridge 源码说明](https://github.com/open-telemetry/opentelemetry-go/blob/main/bridge/opentracing/doc.go)
- **OpenCensus Bridge** 将 OpenCensus 指标和链路接入 OTel SDK。[OpenCensus Bridge 源码说明](https://github.com/open-telemetry/opentelemetry-go/blob/main/bridge/opencensus/doc.go)
- **Log Bridge** 将已有日志库提交的记录接入 OTel 日志流水线，例如将 Go `slog` 日志交给 OTel 日志 SDK。

## 2. Bridge 与 Exporter 的位置

```text
已有埋点或日志库 → Bridge → OTel SDK → Exporter → 后端
```

Bridge 在入口适配埋点调用或数据。SDK 处理这些数据后，由 Exporter 决定怎么把数据提供给下游，例如编码成 OTLP 请求，或将指标转换成 Prometheus 可抓取的表示。
