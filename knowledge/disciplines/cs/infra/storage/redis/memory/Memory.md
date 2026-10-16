---
color: "#5B6CDA"
---

# Memory

Redis 的业务数据主要保存在内存中，容量规划需要同时考虑数据、元数据、客户端缓冲区、复制与持久化开销，以及内存碎片。

## 1. 内存管理

- [内存淘汰](Eviction.md)：达到内存上限后的键选择策略、默认行为及与过期的区别。
- [过期删除](Expiration.md)：TTL 到期后的惰性删除、主动清理与内存释放。

## 2. 容量观察

```redis
INFO memory
INFO stats
```

关注 `used_memory`、`used_memory_rss`、`maxmemory`、`mem_not_counted_for_evict`，以及统计信息中的 `evicted_keys`、`expired_keys` 和命中情况。Redis 统计的分配内存与操作系统看到的进程驻留内存并不完全相同。

内存上限应低于机器或容器限制，并为复制缓冲、持久化、写时复制和其他进程预留空间。
