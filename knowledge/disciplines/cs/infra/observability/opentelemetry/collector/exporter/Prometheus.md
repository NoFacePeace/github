# Prometheus Exporter

Collector 的 Prometheus Exporter 将 `metrics` 流水线中的指标暴露在 Collector 的 `/metrics` 端点，供 Prometheus 抓取。

## 1. 接入指标流水线

例如，应用先通过 OTLP 推送指标。Collector 用 OTLP Receiver 接收指标，再交给 Prometheus Exporter。Prometheus 定期请求 Collector 的 `/metrics`。

```text
应用 ──OTLP 推送──→ Collector 的 OTLP Receiver
                         → Prometheus Exporter → /metrics ← Prometheus 抓取
```
