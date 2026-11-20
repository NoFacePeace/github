---
color: "#5B6CDA"
---

# describeConsumerGroups()

## 1. 查询消费组

传入一个或多个消费组 ID，得到 `DescribeConsumerGroupsResult`。`result.all().get()` 返回以消费组 ID 为键的 `Map<String, ConsumerGroupDescription>`。

```java
DescribeConsumerGroupsResult result = admin.describeConsumerGroups(List.of("group-a"));
ConsumerGroupDescription group = result.all().get().get("group-a");
```

## 2. 返回字段

`ConsumerGroupDescription` 包含消费组状态、协调器和成员列表。每个成员由一个 `MemberDescription` 表示。

| 方法 | 含义 |
| --- | --- |
| `consumerId()` | 消费者在组内的成员 ID。 |
| `clientId()` | 客户端的 `client.id`。 |
| `host()` | 客户端地址。 |
| `groupInstanceId()` | 静态成员 ID，配置 `group.instance.id` 时有值。 |
| `assignment().topicPartitions()` | 当前分配给该成员的 Topic 分区。 |
