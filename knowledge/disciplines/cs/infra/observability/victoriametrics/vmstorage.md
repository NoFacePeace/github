# vmstorage

`vmstorage` 是 VictoriaMetrics 集群的存储节点，保存 `vminsert` 写入的时序样本，并响应 `vmselect` 发来的查询。

## 1. 查询本节点的数据

例如，`vmselect` 请求最近一小时满足 `instance="a"` 的请求计数，`vmstorage` 会按时间范围和标签条件从本节点的数据中找出匹配样本并返回。集群扩容后，原有节点仍可继续保存和提供历史数据。[VictoriaMetrics 集群架构](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#architecture-overview)
