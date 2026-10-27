# 保存派发信息

将已准备的 Job 与状态更新写入 Redis 的[调度记录](../../../../../storage/redis/dispatch-record/DispatchRecord.md)。

## 1. 保存内容

- `state`：变为 `dispatching`。
- `job`：保存包含目标 Runner、租约及授权引用的作业信息。
- `queueWaitMs`：当前时间减去入队时间，表示本次已等待的时长。

## 2. 写入保护

保存前原子检查调度 claim 的令牌。若处理权失效或请求被取消，则不能提交这次派发信息，并回滚本次预留。

保存响应丢失时，也不能直接当作未写入并重新生成作业，因为其他处理流程可能已读取保存结果并继续下发。

## 3. 后续步骤

准备成功后，进入[派发](../dispatch/Dispatch.md)，基于已保存的 Job 下发到 Runner。此时尚未确认下发成功，不能直接标记为 `dispatched` 或认为 Agent 已启动。
