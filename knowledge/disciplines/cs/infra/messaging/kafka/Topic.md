---
color: "#5B6CDA"
---

# Topic

Topic 是 Kafka 中一类消息的命名集合。生产者向指定 Topic 写入消息，消费者通过订阅 Topic 来读取消息。

## 1. 分区组成

一个 Topic 由一个或多个[分区](Partition.md)组成，每条消息写入其中一个分区。例如，`orders` 有 3 个分区，可以分别记作 `orders-0`、`orders-1` 和 `orders-2`。Kafka 的消息顺序以分区为单位。
