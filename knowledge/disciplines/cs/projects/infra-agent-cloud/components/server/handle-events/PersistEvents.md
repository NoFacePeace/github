# 保存事件与结果

Server 将通过执行归属校验的事件写入按任务与执行轮次组织的 Redis 事件存储。

## 1. 普通事件

`accepted`、进度与输出等非终态事件主要更新 Redis 的[执行事件状态](../../../storage/redis/execution-event-state/ExecutionEventState.md)及适用的日志、回放数据。

是否保留供回调消费的输出受回调标记控制。收到 `accepted` 不等于直接将 MySQL 中的 Session 更新为某个接受状态，Session 的主状态仍可能为 `running`。

适用的事件保存在[执行事件回放](../../../storage/redis/ExecutionEventReplay.md)，供持有回调的 Server 读取；原始 Stream 消费确认清理后，回调仍可读取尚未确认的回放成员。

## 2. 终态事件

`completed`、`failed`、`stopped` 等结果进一步更新 [Session](../../../storage/mysql/session/Session.md) 和 [Task](../../../storage/mysql/task/Task.md)，记录状态、结果、退出信息和适用的统计或产物引用。

更新在事务中检查执行状态与已处理序号，避免重复或延迟结果覆盖已结束的执行。

终态处理还涉及执行租约及容量释放，并按规则撤销执行授权。若输出存储超限，可能进入拒绝、停止执行等异常分支，不能按普通事件直接确认成功。

## 3. 后续步骤

处理成功后进入[确认与清理事件](AcknowledgeEvents.md)；失败时通过消费层重试与异常策略处理，不能提前推进为正常成功。
