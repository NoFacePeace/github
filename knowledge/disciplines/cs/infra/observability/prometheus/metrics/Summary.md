# Summary

Summary（摘要）记录观测值的总次数和总和，也可以暴露应用端预先计算的分位数，例如库存查询耗时的 P95。

## 1. 查询分位数

如果应用配置了 P95，可以通过 `quantile` 标签读取这个实例计算的结果。分位数的统计窗口由客户端配置。多个实例各自算出的 P95 无法直接合并成整体 P95。

```promql
inventory_query_duration_seconds{quantile="0.95"}
```
