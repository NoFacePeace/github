# Prometheus

## 指标和 PromQL 的区别

**指标描述采集到的数据，PromQL 描述如何查询、计算这些数据。**

例如，node_exporter 暴露：

```text
node_cpu_seconds_total{cpu="0", mode="idle"} 12345.6
```

表示逻辑 CPU 0 累计空闲了 12345.6 秒，而不是当前 CPU 使用率。

- `node_cpu_seconds_total`：指标名。
- `cpu="0"`、`mode="idle"`：标签，用于区分不同 CPU 和运行模式。
- `12345.6`：当前样本值，单位由指标定义决定。

Prometheus 抓取时通常还会附加 `job`、`instance` 标签。一个指标名与一组完整标签共同标识一条时间序列。

要得到最近 5 分钟 CPU 使用率，需要对累计时间做速率计算，再对核心聚合，具体见 [node_exporter 的 CPU 示例](node-exporter.md#11-cpu)。

## 指标类型与常用 PromQL 语法

| 概念 | 含义 | 使用方式 |
|---|---|---|
| Counter | 累计计数，通常只增不减，重启等情况可能重置 | 用 `rate()` 计算每秒速率，用 `increase()` 计算窗口内增量 |
| Gauge | 当前值，可以增加或减少 | 直接查询或进行比例计算，例如内存容量和 load |
| `[5m]` | 查询时刻之前 5 分钟的样本范围 | 不是采集周期，也不是查询步长 |
| `rate(counter[5m])` | 窗口内平均每秒增长量，处理计数器重置 | 适用于字节数、操作次数、累计耗时等 Counter |
| `avg by (...)` | 按指定标签分组取平均 | 例如对一台主机的多个 CPU 核心取平均 |
| `sum by (...)` | 按指定标签分组求和 | 例如汇总经过筛选的网卡流量 |

对 Counter 应先 `rate()` 再聚合，以便分别识别每条序列的重置。窗口应覆盖多个采集点；15～30 秒抓取周期可先使用 `[5m]`。`increase()` 会做外推，因此结果可能不是整数。函数语义见 [Prometheus 查询函数](https://prometheus.io/docs/prometheus/latest/querying/functions/)。

## 学习文档

- [node_exporter](node-exporter.md)：Linux 主机常用指标含义与 PromQL 示例。
