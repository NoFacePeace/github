# 更新扩展信息

Session 仍处于 `running` 时，尝试将当前调度状态保存为业务侧副本。

## 1. 更新目标

保留 Session 的 [extra_info](../../../../../../storage/mysql/session/ExtraInfo.md) 中已有内容，将调度状态写入其中的 [runnerDispatch](../../../../../../storage/mysql/session/RunnerDispatch.md)。这是更新已有 Session，不新增执行记录。

## 2. 防止状态回退

如果已有调度副本的状态不是 `queued`，延迟到达的排队状态不能将其覆盖回 `queued`。调度信息没有变化时，也无需重复保存。

更新在查询 Session 所在的事务中进行，持有记录写锁，避免并发覆盖。

## 3. 与 Redis 的关系

正常同步只更新 MySQL 中的调度副本，不因此改变 Redis 调度状态。

如果发现 Session 已结束，或存在停止请求，则另行按调度阶段尝试取消排队或发送停止指令。这可能改变 Redis 记录，是后续需要再次读取的原因之一。

## 4. 后续步骤

业务状态同步返回后，进入[再次读取调度记录](RereadDispatchRecord.md)，确认记录是否仍处于排队状态。
