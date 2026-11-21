# vminsert

`vminsert` 是 VictoriaMetrics 集群的写入入口。它接收 Prometheus 或 `vmagent` 等客户端发送的样本，并将样本写到 `vmstorage`。

## 1. 分片与副本

`vminsert` 根据指标名和完整标签集做一致性哈希，选择写入节点。例如，`requests_total{instance="a"}` 和 `requests_total{instance="b"}` 是两条时序，可能被分到不同节点。启用 `-replicationFactor=2` 后，每个新样本会写到两个不同的存储节点。[VictoriaMetrics 集群架构](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#architecture-overview) · [副本机制](https://docs.victoriametrics.com/victoriametrics/cluster-victoriametrics/#replication-and-data-safety)
