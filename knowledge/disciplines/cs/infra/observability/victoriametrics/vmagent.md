# vmagent

`vmagent` 是指标采集和转发程序。它可以定期抓取应用或 Exporter 暴露的指标，再通过远程写入发给 VictoriaMetrics。

## 1. 抓取与转发

例如，可以在每个 Kubernetes 集群部署 `vmagent`，由它抓取本地应用的 `/metrics`，再把指标汇总到同一个 VictoriaMetrics 集群。使用集群版时，写入请求由 `vminsert` 接收。[vmagent 官方文档](https://docs.victoriametrics.com/victoriametrics/vmagent/)
