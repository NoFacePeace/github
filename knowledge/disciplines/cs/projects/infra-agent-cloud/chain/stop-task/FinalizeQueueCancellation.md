# 处理排队取消结果

## 1. 处理内容

取消成功后，释放适用的未发布预留资源，通知调度结果并唤醒队列处理。涉及的共享资源见[执行租约](../../storage/redis/ExecutionLease.md)和 [Runner 容量集合](../../storage/redis/RunnerCapacity.md)。

业务层收到 cancelled 调度结果后，在事务中更新 Session 的扩展信息、状态、退出码和耗时，将仍在 running 的轮次标为 stopped；再按其他运行轮次情况更新 Task。

此分支不需要先启动引擎再停止，也没有 Runner 执行事件序号，应按调度结果直接完成业务状态处理。

## 2. 后续流程

接口侧仍执行[Task 状态更新](UpdateTaskStatus.md)；之后按 Session 实际状态判断是否能继续。
