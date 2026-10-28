# 查询任务与追加模式

## 1. 处理内容

查询已有 [Task](../../storage/mysql/task/Task.md)，确认任务存在且启用 steer 提示词模式。未启用时返回未投递，不将普通继续任务自动当作执行中追加。

该步骤不创建 Task 或 Session，也不重新选择 Agent 默认版本。

## 2. 后续流程

正常路径进入[检查历史投递记录](CheckDeliveryHistory.md)。
