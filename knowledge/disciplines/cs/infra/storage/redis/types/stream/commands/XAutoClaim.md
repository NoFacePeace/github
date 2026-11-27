---
color: "#5B6CDA"
---

# XAUTOCLAIM

`XAUTOCLAIM` 扫描 Stream 消费者组的待确认列表（PEL），将空闲时间达到阈值的消息接管给指定消费者。该命令从 Redis 6.2 起提供，可用于重试消费者掉线后留下的待确认消息。

## 1. 命令语法

```redis
XAUTOCLAIM key group consumer min-idle-time start [COUNT count] [JUSTID]
```

| 参数 | 含义 |
| --- | --- |
| `key`、`group`、`consumer` | Stream 的键、消费者组名和接管消息的消费者名。 |
| `min-idle-time` | 最小空闲时间，单位为毫秒。空闲时间从最近一次投递或接管算起。 |
| `start` | 扫描起点。首轮从 `0-0` 开始，后续使用返回的游标。 |
| `COUNT count` | 本次最多接管的消息数，默认 100。单次最多检查 `count × 10` 条 PEL 记录，因此实际接管数可能少于 `count`。 |
| `JUSTID` | 只返回接管成功的消息 ID，且不增加其投递次数。 |

## 2. 扫描与返回值

```redis
XAUTOCLAIM orders workers consumer-B 60000 0-0 COUNT 10
```

这条命令从 `workers` 组的 PEL 开头扫描，将空闲至少 60 秒的消息交给 `consumer-B`，最多接管 10 条。命令没有 `BLOCK` 参数，扫描结束即返回，无需等待消息变得符合条件。

1. **下一次扫描的游标**。游标为 `0-0` 表示本轮已扫描到末尾。即使本次接管结果为空，只要游标仍有后续位置，就可以用它继续扫描。
2. **成功接管的消息**。默认返回消息 ID 和字段值，加 `JUSTID` 时只返回消息 ID。
3. **本次清理的消息 ID**。Redis 7.0 起，扫描到正文已被删除的消息时，会从 PEL 中清理对应记录，并在这一项返回其 ID。

接管会重置消息的空闲时间。处理成功后，用返回的实际消息 ID 执行 `XACK orders workers <消息ID>`。接管时原消费者可能仍在处理同一条消息，业务处理需要考虑幂等。

## 3. 与 XPENDING 的区别

| 命令 | 用途 | 对 PEL 的影响 |
| --- | --- | --- |
| `XPENDING` | 查看待确认消息的数量，或查询 ID、所属消费者、空闲时间和投递次数。 | 只读。 |
| `XAUTOCLAIM` | 查找并接管达到空闲时间阈值的消息。 | 更新消费者归属并重置空闲时间。 |

排查积压时，可以用 `XPENDING orders workers IDLE 60000 - + 10` 查看空闲至少 60 秒的消息。恢复消费时，可以直接执行 `XAUTOCLAIM`，无需提前查询消息 ID。

[Redis 官方文档](https://redis.io/docs/latest/commands/xautoclaim/)
