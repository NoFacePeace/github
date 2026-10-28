# 尝试取消排队

## 1. 处理内容

读取并原子检查 Redis [调度记录](../../storage/redis/dispatch-record/DispatchRecord.md)，校验 Task 归属。queued 状态可变为 cancelled，清除加密执行输入并删除适用的[调度处理权](../../storage/redis/DispatchClaim.md)；已为 cancelled 时按取消成功处理。

不存在或已不是 queued 时返回未取消，不能把派发中的作业简单当作仍在排队删除。取消成功也不是这一步已经完成所有队列索引清理。

## 2. 后续流程

取消成功进入[处理排队取消结果](FinalizeQueueCancellation.md)；未取消则[检查执行租约](CheckExecutionLease.md)。
