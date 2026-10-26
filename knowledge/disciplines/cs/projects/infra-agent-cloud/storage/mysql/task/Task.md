# Task

Task 表保存一项持续任务的整体信息，例如任务标识、归属、绑定配置和整体状态。

## 1. 生命周期

发起任务时创建 Task 记录；后续继续同一任务时复用该记录。

## 2. 与执行记录的关系

一个 Task 可以包含多个 Session。关联信息由 Session 表的 `task_id` 保存，指向 Task 表的 `id`；Task 无需保存 Session ID 列表。

## 3. Agent 关联字段

- [agent_id](AgentID.md)：任务使用的 Agent。
- [version_id](VersionID.md)：任务绑定的版本记录。
