# 查询已有任务

## 1. 处理内容

按请求中的 UUID 查询已有 [Task](../../storage/mysql/task/Task.md)，沿用其标识、引擎、运行类型及适用的 Agent、版本绑定，不创建新的 Task。

检查任务产物状态是否仍可用。已过期或清理的任务不能仅凭历史记录继续执行；当前链路只讨论远程执行。

## 2. 下一步

进入[检查运行中的执行轮次](CheckRunningSessions.md)。
