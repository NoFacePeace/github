---
color: "#5B6CDA"
---

# Offset

Offset 是记录在 Kafka 分区日志中的位置编号。每个分区独立编号，例如 `orders-0` 和 `orders-1` 上的 `offset=42` 分别指向各自的记录。

## 1. 日志位置

| 名称 | 含义 |
| --- | --- |
| 消息 offset | Kafka 写入记录时为它分配的位置编号。 |
| 日志起始 offset（Log Start Offset） | 分区日志当前保留范围的起点。保留策略删除旧日志后，起点可以向前移动。 |
| 日志末尾 offset（Log End Offset） | 日志下一次追加记录的位置。例如最后一条记录的 offset 是 `49`，则末尾 offset 是 `50`。 |

分区内的 offset 按顺序递增。日志压缩可能删除旧记录并保留其他记录的原 offset，因此两个 offset 的差值可能大于这段范围内仍可读取的消息条数。
