---
color: "#5B6CDA"
---

# XREADGROUP

`XREADGROUP` 以消费者组中某个消费者的身份读取 Stream 消息。默认情况下，新消息投递后会进入该组的待确认列表（PEL），业务处理成功后需要通过 `XACK` 确认。

## 1. 基本用法

```redis
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders >
```

| 参数或选项 | 含义 |
| --- | --- |
| `GROUP workers consumer-A` | 指定消费者组与消费者名称。 |
| `COUNT 10` | 每个 Stream 最多返回 10 条消息。 |
| `STREAMS orders >` | 指定 Stream 及读取位置，此处读取尚未投递给该组的新消息。 |

消费者组必须提前创建；消费者不存在时会自动创建。同一组内的消费者分担新消息，不同组独立消费。

## 2. 读取位置

| 位置 | 读取内容 |
| --- | --- |
| `>` | 尚未投递给该组的新消息。 |
| `0` 或 `0-0` | 当前消费者自己的、ID 大于 `0-0` 的待确认消息。 |
| 具体 ID | 当前消费者自己的、ID 大于指定值的待确认消息。 |

### 2.1 读取新消息

```redis
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders >
```

这里的“新”以消费者组的投递进度为准，不以命令执行时间为准。创建组时若使用 `0`，Stream 中已有但尚未投递给该组的消息也能通过 `>` 读取。

投递后，组的投递进度向前推进，消息默认归当前消费者负责并进入 PEL。其他消费者不会通过 `>` 自动获得这些待确认消息。

### 2.2 重新读取自己的待确认消息

```redis
XREADGROUP GROUP workers consumer-A COUNT 10 STREAMS orders 0
```

该命令不读取组内新消息，也不读取其他消费者负责的待确认消息。分页时传入上次返回的最后一个 ID；每次传 `0` 会再次从自己最早的待确认消息开始读取。

重新投递待确认消息会更新其最近投递时间并增加投递次数。需要接管其他消费者的消息时，应使用 `XCLAIM` 或 `XAUTOCLAIM`。

## 3. 阻塞等待

```redis
XREADGROUP GROUP workers consumer-A COUNT 10 BLOCK 5000 STREAMS orders >
```

没有可投递的新消息时，最多等待 5 秒；有消息可读时立即返回。`BLOCK 0` 表示无限等待。不指定 `BLOCK` 时，没有可读新消息便立即返回空响应。

使用普通 ID 读取自己的待确认消息时，`BLOCK` 会被忽略，不会等待新的待确认消息。这里用于读取新消息的特殊值是 `>`，不能用 `XREAD` 中的 `$` 替代。

## 4. 返回结果

默认按 Stream 返回消息列表，每条消息包含 ID 和 field-value 内容。例如，redis-cli 中的结果为：

```text
1) 1) "orders"
   2) 1) 1) "1000-0"
         2) 1) "order_id"
            2) "1001"
            3) "status"
            4) "created"
```

阻塞超时返回空响应。读取自己的历史待确认消息但没有匹配项时，对应消息列表为空。如果正文已删除而 PEL 记录仍存在，历史读取可能返回消息 ID 和空的正文。

## 5. 确认与 NOACK

默认处理流程是读取、执行业务、确认：

```redis
XACK orders workers 1000-0
```

确认使用实际返回的消息 ID。`XACK` 移除该组的待确认记录，不删除 Stream 中的消息。消费者宕机后，未确认消息仍保留在 PEL，应用需要实现恢复或接管逻辑，并保证业务处理幂等。

```redis
XREADGROUP GROUP workers consumer-A NOACK STREAMS orders >
```

`NOACK` 表示投递新消息时不创建 PEL 记录，适合允许丢失处理结果的场景；消费者读取后宕机时，不能依靠该次投递的待确认记录恢复任务。使用普通 ID 读取历史待确认消息时，`NOACK` 会被忽略。
