---
color: "#5B6CDA"
---

# maxmemory

`maxmemory` 设置 Redis 用于内存限制与淘汰判断的上限。常见 64 位 Redis 默认值为 `0`，表示不设置该上限，不代表机器或容器拥有无限内存。

## 1. 配置方式

```conf
maxmemory 1gb
```

查询和动态修改示例：

```redis
CONFIG GET maxmemory
CONFIG SET maxmemory 1gb
```

运行时修改不会自动保存到配置文件，需要更新部署配置或使用 `CONFIG REWRITE` 保存。

## 2. 达到上限后的行为

行为由 [maxmemory-policy](maxmemory-policy.md) 决定：可以选择淘汰键，也可以保持默认的 `noeviction`，拒绝受内存限制的写入。

降低上限可能引发淘汰或写入失败，修改前应评估当前内存使用和数据用途。

## 3. 实际内存占用

该值不是操作系统看到的进程内存硬上限。部分复制及 AOF 缓冲、内存碎片、后台持久化的写时复制等会影响实际占用，需要低于机器或容器限制并预留余量。
