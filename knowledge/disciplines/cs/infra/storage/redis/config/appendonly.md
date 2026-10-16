---
color: "#5B6CDA"
---

# appendonly

`appendonly` 控制是否开启 AOF 持久化，Redis 默认值为 `no`。实际部署可能覆盖默认值。

## 1. 可选值

| 值 | 含义 |
| --- | --- |
| `yes` | 开启 AOF，记录后续写操作，并建立恢复现有数据所需的基础文件。 |
| `no` | 关闭 AOF，不再追加写操作；不控制 RDB 快照是否启用。 |

## 2. 配置方式

在 `redis.conf` 中开启：

```conf
appendonly yes
```

查询或动态修改当前实例：

```redis
CONFIG GET appendonly
CONFIG SET appendonly yes
```

运行时开启会触发生成包含现有数据的 AOF 基础内容，过程涉及后台持久化工作。应确认生成成功后，再将 AOF 作为有效的恢复来源。

运行时修改不会自动写回配置文件，需要更新部署配置或使用 `CONFIG REWRITE` 保存。

## 3. 同步策略

开启后，通过 [appendfsync](appendfsync.md) 控制同步到磁盘的策略，默认值为 `everysec`。
