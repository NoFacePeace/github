# 确认调度结果

将已保存的调度结果同步到业务处理层，按当前 Session 状态更新 [runnerDispatch](../../../../../storage/mysql/session/RunnerDispatch.md) 或执行相应收尾。

## 1. 同步成功后确认

业务同步确认成功后，在校验调度 claim 的条件下，通过同一 Redis Lua 脚本完成：

- 从[待调度集合](../../../../../storage/redis/DispatchPending.md)移除本轮 Session。
- 为对应调度记录设置 24 小时 TTL。

排队中或派发中记录不能按已完成调度确认。同步失败或结果尚未满足确认条件时，保留记录供后续恢复。

## 2. 与执行完成的区别

调度结果确认结束的是本次排队与派发处理。Runner 后续消费 Job、启动 Agent 和回传执行结果，属于接下来的执行链路。
