# 保存派发结果

投递步骤成功后，将 Redis [调度记录](../../../../../storage/redis/dispatch-record/DispatchRecord.md) 的状态改为 `dispatched`，并校验调度 claim 令牌后保存。

## 1. 失败分支

执行归属冲突、投递冲突或租约截止时间已过等明确错误，按对应逻辑回滚适用的预留并保存 `failed` 状态。其他异常可能需要后续恢复重试，不能一概视为消息未投递。

## 2. 后续步骤

保存成功后，进入[确认调度结果](AcknowledgeDispatch.md)。`dispatched` 只表示下发成功，不是 Session 执行完成状态。
