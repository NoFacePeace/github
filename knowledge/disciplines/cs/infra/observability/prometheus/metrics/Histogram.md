# Histogram

Histogram（直方图）用桶统计多次观测值的分布，例如库存查询耗时。以 Prometheus 的经典直方图为例，它会暴露累计桶计数、观测总次数和观测值总和。

## 1. 累计桶与分位数

如果三次查询分别耗时 0.05 秒、0.2 秒和 0.6 秒，那么 `le="0.1"` 的桶计数为 1，`le="0.5"` 的桶计数为 2。每个桶统计小于或等于该上界的观测次数。

多个实例使用相同的桶边界时，可以按 `le` 合并桶，再用 `histogram_quantile` 估算整体 P95。

```promql
histogram_quantile(0.95, sum by (le) (rate(inventory_query_duration_seconds_bucket[5m])))
```
