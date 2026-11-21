# vmselect

`vmselect` 是 VictoriaMetrics 集群的查询入口。它接收 Grafana 等客户端发来的查询，再向所有已配置的 `vmstorage` 节点获取匹配数据。

## 1. 汇总查询结果

例如，查询所有实例的请求计数时，相关时序可能分散在多个存储节点上。`vmselect` 会收集各节点返回的数据，再计算并返回查询结果。启用应用层副本时，可通过 `-dedup.minScrapeInterval=1ms` 去重同一时序中时间戳相同的副本样本。[VictoriaMetrics 集群架构](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#architecture-overview) · [副本机制](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#replication-and-data-safety)
