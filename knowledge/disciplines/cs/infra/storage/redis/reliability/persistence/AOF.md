---
color: "#5B6CDA"
---

# AOF

AOF（Append Only File，追加日志文件）是 Redis 的一种持久化方式，通过记录修改数据的操作，在重启时恢复数据。`GET` 等只读命令不会记录。

## 1. 数据恢复

例如执行以下命令：

```redis
SET count 1
INCR count
```

AOF 记录相应的写操作，重启时依次重放，恢复出 `count = 2`。日志中的命令可能经过规范化，并不一定与客户端发送的命令逐字相同。

AOF 整体包含恢复完整数据集所需的信息，但日常写入只追加增量操作。恢复需要完整的文件集合，单独使用最近一段增量日志通常无法恢复全部数据。

## 2. 默认配置

Redis 通过 [appendonly](../../config/appendonly.md) 控制是否开启 AOF，默认不开启；通过 [appendfsync](../../config/appendfsync.md) 设置同步策略。默认配置为：

```conf
appendonly no
appendfsync everysec
```

启用 AOF 后，默认采用 `everysec` 同步策略。实际部署可能覆盖默认配置，可查询当前实例：

```redis
CONFIG GET appendonly
CONFIG GET appendfsync
```

## 3. 同步策略

写入文件与同步到磁盘是两个步骤。写入可能仅到达操作系统的页缓存，[appendfsync](../../config/appendfsync.md) 控制何时执行 `fsync`，将数据同步到磁盘。

| 策略 | 同步行为 | 主要取舍 |
| --- | --- | --- |
| `always` | 每批 AOF 写入都同步到磁盘 | 持久性更强，同步开销更大。 |
| `everysec` | 通常每秒同步一次 | 性能与持久性的常用折中，故障时通常可能丢失约 1 秒的数据。 |
| `no` | 不主动执行 `fsync`，由操作系统决定何时同步 | 同步开销较低，数据丢失窗口更不确定。 |

实际数据丢失范围还受磁盘延迟、操作系统和硬件故障等因素影响，`everysec` 并不是严格的 1 秒丢失上限。

## 4. 日志重写

随着写操作积累，AOF 会持续增长。[重写](Rewrite.md)根据当前全量数据生成更紧凑的恢复记录，避免保留全部历史操作。例如计数器更新一万次，重写后只需保存恢复其当前值所需的记录。

Redis 7 起采用多文件 AOF，由基础文件、增量文件及记录文件集合的 manifest 组成。基础文件保存重写时的数据状态，增量文件记录后续变更；开启 `aof-use-rdb-preamble` 时，基础文件使用 RDB 格式，这种模式称为[混合持久化](HybridPersistence.md)。

## 5. 与 RDB 的区别

RDB 保存某个时间点的数据快照；AOF 在基础数据之上持续记录写操作，通常可以提供更小的数据丢失窗口。两者都启用时，Redis 重启优先使用 AOF 恢复数据。

AOF 文件也可能因磁盘故障或误删除而丢失，需要额外备份并验证恢复流程。
