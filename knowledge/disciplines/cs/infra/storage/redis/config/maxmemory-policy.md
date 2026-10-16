---
color: "#5B6CDA"
---

# maxmemory-policy

`maxmemory-policy` 决定达到 `maxmemory` 限制时是否淘汰键，以及如何选择候选键。默认值为 `noeviction`。

## 1. 可选值

| 值 | 行为 |
| --- | --- |
| `noeviction` | 不淘汰，拒绝受内存限制的写入。 |
| `allkeys-lru` | 在所有键中按近似 LRU 淘汰。 |
| `allkeys-lfu` | 在所有键中按近似 LFU 淘汰。 |
| `allkeys-random` | 在所有键中随机淘汰。 |
| `volatile-lru` | 在设置了过期时间的键中按近似 LRU 淘汰。 |
| `volatile-lfu` | 在设置了过期时间的键中按近似 LFU 淘汰。 |
| `volatile-random` | 在设置了过期时间的键中随机淘汰。 |
| `volatile-ttl` | 在设置了过期时间的键中优先淘汰剩余 TTL 较短的键。 |

`volatile-*` 不要求键已经过期；候选键不足时仍可能写入失败。LRU 关注最近访问时间，LFU 关注访问频率，具体机制和使用边界见[内存淘汰](../memory/Eviction.md)。

## 2. 配置方式

```conf
maxmemory-policy allkeys-lru
```

查询和动态修改示例：

```redis
CONFIG GET maxmemory-policy
CONFIG SET maxmemory-policy allkeys-lru
```

需要配合非零的 [maxmemory](maxmemory.md) 上限才会按内存压力触发淘汰。运行时修改不会自动保存到配置文件，需要更新部署配置或使用 `CONFIG REWRITE` 保存。

## 3. 数据用途

缓存通常可以容忍键被提前淘汰；锁、消息和业务事实数据应单独评估。策略影响整个实例的候选键选择，不能仅凭键名前缀排除某类数据。
