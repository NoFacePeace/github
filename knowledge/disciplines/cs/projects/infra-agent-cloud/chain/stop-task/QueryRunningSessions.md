# 查询运行中的 Session

## 1. 处理内容

按 UUID 查询已有 [Task](../../storage/mysql/task/Task.md)，检查其允许由 Server 停止，再查询该 Task 下所有 running Session，逐个请求停止。

若没有 running Session 且 Task 已非 running，可直接返回已有状态；若有目标 Session，则进入停止处理。

接口层完成目标轮次停止请求后，还会独立[更新 Task 状态](UpdateTaskStatus.md)。这不要求目标 Session 已完成终态落库。

## 2. 后续流程

[写入停止请求标记](MarkStopRequested.md)。
