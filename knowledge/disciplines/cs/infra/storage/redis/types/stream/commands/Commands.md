---
color: "#5B6CDA"
---

# Commands

Stream 命令用于追加消息、按 ID 读取、管理消费者组、确认消息及清理历史记录。

## 1. 命令概览

| 命令 | 作用 |
| --- | --- |
| `XADD` | 追加消息，返回消息 ID。 |
| `XLEN` | 查询 Stream 中的消息条数。 |
| `XRANGE`、`XREVRANGE` | 按 ID 范围正向或反向查询消息。 |
| `XREAD` | 从指定 ID 之后读取消息，支持阻塞等待。 |
| [XGROUP](xgroup/XGroup.md) | 管理消费者组和消费者，创建组并设置起始投递位置。 |
| [XREADGROUP](XReadGroup.md) | 读取组内新消息或当前消费者的待确认消息。 |
| [XACK](XAck.md) | 确认处理完成，移除组内待确认记录。 |
| [XPENDING](XPending.md) | 查询消费者组的待确认消息。 |
| [XCLAIM](XClaim.md)、`XAUTOCLAIM` | 接管满足空闲时间条件的待确认消息。 |
| `XDEL` | 删除指定 ID 的消息。 |
| `XTRIM` | 按长度或最小 ID 裁剪消息。 |
| `XINFO` | 查看 Stream、消费者组和消费者信息。 |

## 2. 追加与范围查询

```redis
XADD orders * order_id 1001 status created
XLEN orders
XRANGE orders - + COUNT 10
```

`*` 表示自动生成递增的消息 ID，通常采用 `毫秒时间戳-序号` 格式。消息内容是一组 field-value 对，字段和值都是字符串，不是独立的 Redis Hash。

`XRANGE` 的起止 ID 默认都包含在范围内；`-` 和 `+` 分别表示最小和最大边界。

## 3. 普通读取

```redis
XREAD COUNT 10 STREAMS orders 0-0
XREAD BLOCK 5000 STREAMS orders $
```

`XREAD` 返回 ID 大于指定位置的消息。`0-0` 用于从头读取；`$` 表示以命令执行时的末尾为起点，等待之后到达的新消息。持续读取时，应传入上次返回的最后一个 ID，避免每次使用 `$` 跳过期间到达的消息。

`BLOCK 5000` 表示最多等待 5 秒。普通读取不记录确认状态，客户端需要自行保存读取进度。

## 4. 消费者组与确认

```redis
XGROUP CREATE orders workers 0 MKSTREAM
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders >
```

创建组时，`0` 表示从已有消息开始消费，`$` 表示只消费创建组之后的新消息；`MKSTREAM` 会在 key 不存在时创建空 Stream。

`>` 用于读取尚未投递给组内消费者的新消息。默认情况下，消息投递后进入该组的待确认列表（PEL）。业务处理成功后，使用返回的实际消息 ID 确认：

```redis
XACK orders workers 1730000000000-0
```

`XACK` 返回成功移除的待确认记录数，不删除 Stream 中的消息。

## 5. 待确认消息与接管

```redis
XPENDING orders workers
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders 0-0
XAUTOCLAIM orders workers consumer-B 60000 0-0 COUNT 10
```

`XPENDING` 查看组内待确认情况。`XREADGROUP` 使用 `0-0` 时，读取当前消费者自己名下、ID 大于该位置的待确认消息，不读取其他消费者的待确认消息，也不读取组内新消息。

`XAUTOCLAIM` 从指定扫描位置开始，将空闲至少 60 秒的待确认消息接管给 `consumer-B`，并返回消息内容和后续扫描游标。继续扫描时使用返回的游标；返回 `0-0` 表示本轮扫描结束。该命令从 Redis 6.2 起提供。

应用需要实现接管、重试及幂等处理，避免重复消费造成重复业务操作。

## 6. 删除与裁剪

```redis
XDEL orders 1730000000000-0
XTRIM orders MAXLEN ~ 10000
XADD orders MAXLEN ~ 10000 * order_id 1002 status created
```

`XDEL` 删除指定消息；`XTRIM` 裁剪已有消息；`XADD` 可以在追加时执行裁剪。`MAXLEN ~ 10000` 表示近似保留最新 1 万条，实际条数可能超过该值。

消费和确认不会自动清理消息。默认删除或裁剪不会等待所有消费者处理完成，也不会同步移除所有相关 PEL 记录，保留策略需要考虑消费进度。
