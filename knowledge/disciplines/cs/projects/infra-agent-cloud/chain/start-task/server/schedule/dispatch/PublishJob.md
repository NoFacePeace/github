# 投递 Job

向目标 Runner 的 [Job Stream](../../../../../storage/redis/JobStream.md) 追加作业消息，并通过[派发去重标记](../../../../../storage/redis/PublishedJob.md)防止重试重复写入。

## 1. 原子投递

在同一 Redis Lua 脚本中执行：

1. 检查去重标记；相同消息已发布则直接返回，不同消息则报告冲突。
2. 未发布时，按 Redis 当前时间检查 Job 中的租约截止时间，过期则拒绝投递。
3. 通过 `XADD` 将序列化 Job 写入 Stream 的 `payload` 字段。
4. 保存去重标记并设置有效期。

此处校验的是 Job 中的租约截止时间，不应理解为重新执行了完整的 Runner 准入过程。

## 2. 后续步骤

投递成功或确认同一消息已投递后，进入[保存派发结果](SaveDispatchResult.md)。网络错误等不明确结果应交由恢复流程处理，不能直接生成新 Job 重发。
