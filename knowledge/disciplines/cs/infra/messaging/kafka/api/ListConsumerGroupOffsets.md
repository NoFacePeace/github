---
color: "#5B6CDA"
---

# listConsumerGroupOffsets()

## 1. 查询已提交 offset

传入消费组 ID，得到 `ListConsumerGroupOffsetsResult`。查询单个消费组时，`partitionsToOffsetAndMetadata().get()` 返回按 `TopicPartition` 索引的 `OffsetAndMetadata`。调用 `offset()` 可读取已提交的位置。

```java
Map<TopicPartition, OffsetAndMetadata> committed = admin.listConsumerGroupOffsets("group-a")
    .partitionsToOffsetAndMetadata().get();
```

结果中的 offset 是消费组在相应分区的恢复位置。计算 lag 时，还需要用 [`listOffsets()`](ListOffsets.md) 查询同一分区的末尾位置，再减去已提交 offset。
