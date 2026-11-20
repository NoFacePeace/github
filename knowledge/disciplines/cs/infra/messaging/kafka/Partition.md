---
color: "#5B6CDA"
---

# Partition

Partition（分区）是 Kafka 存储消息和分配消费任务的单位。分区编号在所属 Topic 内标识分区，例如 `orders-0` 表示 `orders` 的 0 号分区。

## 1. 分区日志

每个分区是一条有序的追加日志。Kafka 按写入顺序为消息分配递增的 [Offset](Offset.md)。一条消息的位置可以用 Topic、分区编号和 offset 一起表示。

## 2. 分区与并行消费

同一 Topic 的不同分区可以并行读取。使用 Kafka 的订阅式消费组时，同一组内每个分区同一时间最多分配给一个成员。
