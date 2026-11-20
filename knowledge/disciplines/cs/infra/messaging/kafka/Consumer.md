---
color: "#5B6CDA"
---

# Consumer

Consumer（消费者）是从 Kafka 分区读取消息的客户端实例。使用 Java 客户端时，可以用 `subscribe()` 订阅 Topic 并作为消费组成员获取分区分配，也可以用 `assign()` 手动指定分区。应用调用 `poll()` 获取记录，再处理返回的消息。

## 1. 当前读取位置

消费者实例为每个已分配的分区维护一个当前读取位置（position），表示下次读取的位置。`poll()` 返回记录时会推进这个位置，可以用 `position(TopicPartition)` 查询。

例如，消费者 A 从 `orders-0` 读取到 `offset=49` 后，下次读取位置可以是 `50`。如果消费组最后一次提交的是 `43`，实例的读取位置和消费组的恢复位置分别是 `50` 和 `43`。

## 2. 消费组成员详情

通过 [`describeConsumerGroups()`](api/DescribeConsumerGroups.md) 可以查询消费者的成员 ID、客户端地址和当前分区分配。
