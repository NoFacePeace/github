# 调度本轮执行

## 1. 处理内容

调度器读取本轮排队请求，检查状态与超时、准备 Runner 和授权、发布 Job，再确认调度完成。

该阶段共用[调度流程](../start-task/server/schedule/Schedule.md)。本轮使用新的 Session 标识与执行租约；调度准备的恢复标识来自已有引擎会话信息。

## 2. 下一步

进入[恢复文件并准备环境](RestoreExecution.md)。
