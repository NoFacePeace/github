---
color: "#5B6CDA"
---

# CREATE

`XGROUP CREATE` 在指定 Stream 上创建消费者组，并设置该组的初始投递位置。

## 1. 基本用法

```redis
XGROUP CREATE orders workers 0 MKSTREAM
```

| 参数 | 含义 |
| --- | --- |
| `orders` | Stream 的 key。 |
| `workers` | 要创建的消费者组名称。 |
| `0` | 将初始投递位置设为 `0-0`，后续从最早保留的消息开始投递。 |
| [MKSTREAM](MKStream.md) | 可选选项，key 不存在时，先创建空 Stream。 |

创建成功返回 `OK`。同一个 Stream 中已存在同名组时，返回 `BUSYGROUP` 错误。不带 `MKSTREAM` 时，目标 Stream 必须已存在。

Stream 保存消息，消费者组保存消费状态。可以先通过 `XADD` 创建 Stream，再创建组；也可以通过 `CREATE ... MKSTREAM` 在创建组时顺便创建空 Stream。

## 2. 初始投递位置

| 位置 | 后续使用 `XREADGROUP ... >` 时的行为 |
| --- | --- |
| `0` 或 `0-0` | 从最早保留的消息开始投递。 |
| `$` | 跳过当前已有消息，只投递之后追加的新消息。 |
| 具体消息 ID | 从大于该 ID 的消息开始投递。 |

该位置表示组的初始投递边界，不表示消息已经处理完成。跳过的消息没有进入该组的 PEL，也不等于执行了 `XACK`；其正文仍保留在 Stream 中，直到被删除或裁剪。

## 3. 创建后读取

```redis
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders >
```

`consumer-A` 是组内消费者名称，首次读取时可以自动创建，无需提前执行 `CREATECONSUMER`。默认情况下，新消息投递后进入 PEL，业务处理成功后再执行 `XACK`。
