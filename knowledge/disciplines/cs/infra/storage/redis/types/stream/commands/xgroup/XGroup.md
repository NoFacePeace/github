---
color: "#5B6CDA"
---

# XGROUP

`XGROUP` 用于管理 Stream 的消费者组和消费者，包括创建组、调整投递位置、显式创建消费者以及删除消费者或组。读取消息使用 `XREADGROUP`，确认消息使用 `XACK`。

## 1. 子命令概览

| 子命令 | 作用 |
| --- | --- |
| [CREATE](Create.md) | 创建消费者组，设置初始投递位置。 |
| `SETID` | 修改已有消费者组的投递位置。 |
| `CREATECONSUMER` | 显式创建组内消费者。 |
| `DELCONSUMER` | 删除组内消费者及其待确认记录。 |
| `DESTROY` | 删除消费者组及其消费状态。 |
| `HELP` | 查看子命令帮助。 |

## 2. 创建消费者组

使用 `XGROUP CREATE` 创建组并设置初始投递位置，可通过 `MKSTREAM` 在目标 key 不存在时创建空 Stream。

## 3. 修改投递位置

```redis
XGROUP SETID orders workers 0
XGROUP SETID orders workers $
```

两条命令分别将组的投递位置移到开头和当前末尾，成功返回 `OK`。这是两个独立示例，应按需要选择执行。

`SETID` 影响后续 `XREADGROUP ... >` 的读取位置，不会清空或确认已有 PEL。向前调整可能重新投递之前处理过的消息，向后调整可能跳过尚未投递的消息。

## 4. 管理消费者

### 4.1 显式创建

```redis
XGROUP CREATECONSUMER orders workers consumer-A
```

新建消费者返回 `1`，消费者已存在返回 `0`。组必须已存在。该子命令从 Redis 6.2 起提供。

### 4.2 删除消费者

```redis
XGROUP DELCONSUMER orders workers consumer-A
```

删除消费者，返回该消费者原有的待确认消息数量。它不会将这些消息自动分配给其他消费者；删除后，这些消息不能再通过原有 PEL 记录接管。

需要保留待确认任务时，应先处理并确认消息，或将消息接管给其他消费者，再删除原消费者。删除消费者不会删除 Stream 中的消息正文。

## 5. 删除消费者组

```redis
XGROUP DESTROY orders workers
```

删除组及其消费者、投递进度和 PEL，成功删除返回 `1`，组不存在返回 `0`。Stream 本身及其中的消息仍保留。

删除后重新创建同名组会建立新的消费状态，不会恢复原组的投递进度和待确认记录。
