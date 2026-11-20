---
color: "#5B6CDA"
---

# listOffsets()

## 1. 查询分区 offset

传入 `Map<TopicPartition, OffsetSpec>`，为每个分区指定要查询的位置。`OffsetSpec.earliest()` 查询起始位置，`OffsetSpec.latest()` 查询末尾位置。

## 2. 返回结果

方法返回 `ListOffsetsResult`。`all().get()` 返回以 `TopicPartition` 为键的结果，调用 `ListOffsetsResultInfo.offset()` 读取 offset。

```java
TopicPartition partition = new TopicPartition("orders", 0);
long endOffset = admin.listOffsets(Map.of(partition, OffsetSpec.latest()))
    .all().get().get(partition).offset();
```
