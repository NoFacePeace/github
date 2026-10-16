---
color: "#5B6CDA"
---

# appendfsync

`appendfsync` 控制 AOF 执行 `fsync`、将数据从操作系统页缓存同步到磁盘的策略。Redis 默认值为 `everysec`，在 [appendonly](appendonly.md) 开启时生效。

## 1. 可选值

| 值 | 同步行为 | 主要取舍 |
| --- | --- | --- |
| [always](appendfsync/always.md) | 每批 AOF 写入都同步到磁盘 | 持久性更强，同步开销更大。 |
| [everysec](appendfsync/everysec.md) | 通常每秒同步一次 | 性能与持久性的常用折中，故障时通常可能丢失约 1 秒的数据。 |
| [no](appendfsync/no.md) | 不主动执行 `fsync`，由操作系统决定何时同步 | 同步开销较低，数据丢失窗口更不确定。 |

`no` 不代表关闭 AOF，只是不主动同步到磁盘。实际数据丢失范围还受磁盘延迟、操作系统和硬件故障等因素影响，`everysec` 并不是严格的 1 秒丢失上限。

## 2. 配置方式

在 `redis.conf` 中设置：

```conf
appendfsync everysec
```

查询或动态修改当前实例：

```redis
CONFIG GET appendfsync
CONFIG SET appendfsync everysec
```

运行时修改不会自动写回配置文件，需要更新部署配置或使用 `CONFIG REWRITE` 保存。
