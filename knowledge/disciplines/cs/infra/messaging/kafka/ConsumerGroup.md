---
color: "#5B6CDA"
---

# Consumer Group

Consumer Group（消费组）由共享 `group.id` 并协同读取分区的[消费者](Consumer.md)组成。不同消费组可以独立读取同一分区，并各自保存消费进度。

## 1. 分区分配

使用订阅式消费组时，同一组内一个分区同一时间最多分配给一个消费者，一个消费者可以负责多个分区。例如，4 个消费者在同一组内都只订阅有 3 个分区的 `orders`，最多有 3 个消费者会获得分区分配。成员加入或退出时，Kafka 可以重新分配分区。

## 2. 消费进度

Kafka 按 `group.id + Topic + 分区` 保存已提交 offset（committed offset）。消费者接管分区时，从有效的已提交 offset 恢复。

例如，`group-a` 的消费者 A 已经读完 `orders-0` 的 `offset=49`，此时它的 position 是 `50`。如果它最后只提交了 `43`，消费者 B 接管分区后会从有效的已提交位置 `43` 恢复，因此可能重复读取 `43` 到 `49`。消费组的已提交进度与日志末尾之间的差值称为 [Lag](Lag.md)。

## 3. 查询消费组

### 3.1 命令行

使用 Kafka 官方工具 [kafka-consumer-groups.sh](scripts/kafka-consumer-groups.md) 查询分区进度、lag 和成员分配。

### 3.2 Java Admin API

| 方法 | 查询结果 |
| --- | --- |
| [`describeConsumerGroups()`](api/DescribeConsumerGroups.md) | 消费组状态、协调器和成员详情。 |
| [`listConsumerGroupOffsets()`](api/ListConsumerGroupOffsets.md) | 消费组在各分区上的已提交 offset。 |
| [`listOffsets()`](api/ListOffsets.md) | 指定分区的 offset，可查询分区末尾。 |

计算消费组 lag 时，按 `TopicPartition` 对齐已提交 offset 和分区末尾 offset，再计算差值。
