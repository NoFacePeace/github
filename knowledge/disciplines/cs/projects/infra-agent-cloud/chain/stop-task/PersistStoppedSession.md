# 保存 Session 停止终态

## 1. 处理内容

Server 消费执行事件并校验执行归属、序号与适用终态后，在事务中更新 [Session](../../storage/mysql/session/Session.md) 的状态、退出信息、耗时、引擎会话标识及适用结果。

当前轮次存在 runnerStopRequested 标记时，业务状态按 stopped 处理。随后更新适用的 Task 状态并处理共享租约、容量和授权等资源，具体共用流程见 [Server 事件处理](../../components/server/handle-events/HandleEvents.md)。

Session 实际退出 running 后，才满足 Continue 的“没有运行轮次”检查；停止本身不会创建新 Session。

## 2. 后续流程

继续任务还要求任务产物可用、配置与授权检查通过。
