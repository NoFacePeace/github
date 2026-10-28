---
color: "#5B6CDA"
---

# ID

## 1. 格式与作用

消息 ID 在同一个 Stream 内唯一，追加消息时必须严格递增。不同 Stream 可以出现相同 ID。

ID 采用两个无符号 64 位整数构成的 `毫秒部分-序号` 格式。比较大小时，先比较毫秒部分，再比较序号：

```text
1-0 < 1-1 < 2-0
```

消息 ID 用于定位消息、指定读取位置、确认和删除消息。它与消息内容中的 `order_id` 等业务字段相互独立。

消息 ID 参数中 `0` 与 `0-0` 的关系及其在不同命令中的作用见 [0](Zero.md)。

## 2. 自动生成

```redis
XADD orders * order_id 1001
```

Redis 以当前毫秒时间为基础生成 ID；同一毫秒内通过递增序号区分消息。如果当前时间小于 Stream 已记录的最大 ID 的毫秒部分，会沿用该毫秒部分并递增序号，以维持 ID 递增。

## 3. 手动指定

以下示例使用一个尚不存在的 Stream：

```redis
XADD manual-orders 1-0 order_id 1001
XADD manual-orders 1-1 order_id 1002
XADD manual-orders 2-0 order_id 1003
```

手动指定时，毫秒部分不要求对应真实时间，可以使用业务递增编号。指定的 ID 必须大于 `0-0`，并严格大于该 Stream 已生成的最大 ID。

如果此前使用 `*` 生成了基于当前时间的 ID，之后再使用较小的业务编号会报错。

Redis 7.0 起，支持指定毫秒部分并自动生成序号：

```redis
XADD manual-orders 2-* order_id 1004
```

## 4. 重复与倒退

```redis
XADD manual-orders 2-0 order_id 1005
```

如果 Stream 已生成过 `2-0` 或更大的 ID，上述命令会报错，消息不会写入，也不会覆盖已有消息。

删除最大 ID 对应的消息后，Stream 仍保留已生成的最大 ID，不能通过删除消息来重复使用旧 ID。这与删除整个 Stream key 后重新创建不同。
