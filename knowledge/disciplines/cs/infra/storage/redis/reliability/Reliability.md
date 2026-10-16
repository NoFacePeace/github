---
color: "#5B6CDA"
---

# Reliability

Redis 可靠性关注故障时的数据保留、恢复与服务可用性。

## 1. 持久化

- [持久化](persistence/Persistence.md)：通过 RDB 快照和 AOF 日志保存数据，以支持重启或故障后的恢复。

## 2. 主从复制

- [主从复制](Replication.md)：数据副本、全量与部分同步，以及复制延迟和数据丢失边界。

## 3. 哨兵

- [Sentinel](Sentinel.md)：监控主从节点、协调自动故障切换，并向客户端提供主节点发现。
