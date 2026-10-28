# 确保消费组存在

消费循环开始前，为资源池公共流和当前 Runner 的目标流确保消费组存在。目标流的结构见 [Job Stream](../../../../storage/redis/JobStream.md)，公共流采用相同前缀但不带末尾的 Runner ID。

## 1. 初始化

```text
XGROUP CREATE <stream> <group> 0 MKSTREAM
```

Stream 不存在时创建；消费组已存在则复用，不在每轮循环中重新创建。起始 ID 为 `0`，新建组可以读取已有历史消息。

消费组与 Pending 列表属于 Stream 内部状态，不是独立 Redis key。

## 2. 后续步骤

进入消费循环，确认处理并发有空位后，开始[接管未确认消息](ClaimPendingMessages.md)。处理数达到上限时，先等待已有处理返回。
