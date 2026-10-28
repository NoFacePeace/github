# 更新任务状态

## 1. 处理内容

将已有 [Task](../../storage/mysql/task/Task.md) 的状态更新为 running，并更新内存中的 Task 对象。

此次操作只更新已有 Task；新 Session 已在前一步创建。Task 和 Session 的创建、更新不是在这里一次性完成的同一条写入。

## 2. 下一步

进入[异步提交远程执行](SubmitExecution.md)。
