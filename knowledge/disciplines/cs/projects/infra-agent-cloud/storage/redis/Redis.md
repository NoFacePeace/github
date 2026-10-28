# Redis

Redis 是内存键值存储，可在 Agent 云平台中承载缓存、任务队列、租约、心跳和执行协调状态。

## 1. 工具上下文

- [内部工具上下文](InternalToolContext.md)：按上下文标识保存访问令牌哈希和工具运行资料。

## 2. 执行事件

- [执行事件状态](execution-event-state/ExecutionEventState.md)：按任务与执行轮次保存回调标记等事件处理状态。
- [执行事件 Stream](RunnerEventStream.md)：保存 Runner 回传、供 Server 消费的事件。
- [执行事件回放](ExecutionEventReplay.md)：保存按执行轮次组织、供本地回调读取的事件。
- [执行日志](ExecutionLog.md)：保存从输出事件提取的引擎日志，供历史查询与对话解析。
- [事件字节计数](RunnerEventBytes.md)：统计事件 payload 总大小，控制积压。

## 3. 调度

- [调度记录](dispatch-record/DispatchRecord.md)：保存单轮执行的排队状态与加密输入。
- [调度处理权](DispatchClaim.md)：通过随机令牌和短期 TTL 协调单条调度记录的处理。
- [待调度集合](DispatchPending.md)：按排队时间组织等待调度的 Session。
- [调度通知频道](DispatchNotification.md)：发布唤醒通知，提示调度器检查队列。

## 4. 执行准入

- [执行租约](ExecutionLease.md)：记录本轮执行的 Runner 归属与租约标识。
- [Runner 容量集合](RunnerCapacity.md)：记录各节点已预留的租约名额。
- [Runner 维护标记](RunnerMaintenance.md)：记录停止接单的维护要求，恢复时显式清除。

## 5. 作业投递

- [Job Stream](JobStream.md)：保存投递给目标 Runner 的作业消息。
- [派发去重标记](PublishedJob.md)：记录已投递消息，避免重试重复追加。
