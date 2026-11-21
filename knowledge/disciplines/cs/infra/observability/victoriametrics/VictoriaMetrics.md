# VictoriaMetrics

VictoriaMetrics（VM）是用于存储和查询指标的时序数据库。它有单机版和集群版，使用兼容大多数 PromQL 查询的 MetricsQL。

## 1. 与 Prometheus 的分工

Prometheus 可以独立抓取指标、存储数据、执行查询和计算告警规则，每个服务端实例使用自己的本地时序数据库。VictoriaMetrics 单机版也能直接抓取指标，集群版则可以分片存储，并集中接收多个采集端的数据。[Prometheus 架构](https://prometheus.io/docs/introduction/overview/) · [VictoriaMetrics 单机版](https://docs.victoriametrics.com/victoriametrics/single-server-victoriametrics/)

例如，可以让各 Kubernetes 集群的 Prometheus 负责本地抓取和告警，再通过 `remote_write` 把指标写到同一个 VictoriaMetrics。Grafana 可以查询 VictoriaMetrics 中汇总的历史指标。也可以由 [vmagent](vmagent.md) 负责抓取并转发。

## 2. 集群如何写入和查询

| 组件 | 职责 |
| --- | --- |
| [vminsert](vminsert.md) | 根据指标名和完整标签集做一致性哈希，选择写入的存储节点 |
| [vmstorage](vmstorage.md) | 存储数据，按时间范围和标签条件返回匹配数据 |
| [vmselect](vmselect.md) | 向所有已配置的存储节点查询，再汇总结果 |

例如，`requests_total{instance="a"}` 和 `requests_total{instance="b"}` 是两条时序，可能写到不同节点。查询 `requests_total` 时，`vmselect` 会从各节点收集匹配结果。扩容后，历史数据仍可留在原节点，只要查询侧配置了新旧节点，就可以一起查询。[VictoriaMetrics 集群架构](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#architecture-overview)

## 3. 多副本

在 `vminsert` 上配置 `-replicationFactor=2` 后，每个新样本会写到两个不同的 `vmstorage` 节点。查询端的 `vmselect` 用 `-dedup.minScrapeInterval=1ms` 对这些副本去重。

以三个存储节点为例，一个节点下线后，已成功保存两份的样本仍然可以从剩余节点读到。剩余两个节点也可以继续给新样本各存一份，维持双副本。[VictoriaMetrics 副本机制](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#replication-and-data-safety)
