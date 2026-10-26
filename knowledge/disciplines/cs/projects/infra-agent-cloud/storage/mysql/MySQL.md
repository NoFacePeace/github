# MySQL

MySQL 是关系型数据库，可在 Agent 云平台中持久化任务、执行记录和配置等结构化业务数据。

## 1. Agent 与版本

- [Agent](agent/Agent.md)：保存智能体身份、状态和发布信息。
- [AgentVersion](agent-version/AgentVersion.md)：保存各版本的名称、描述及执行配置。

一个 Agent 可以包含多个版本，版本记录通过 `agent_id` 关联 Agent。

## 2. 任务与执行记录

- [Task](task/Task.md)：保存一项持续任务的整体信息。
- [Session](session/Session.md)：保存任务中一轮执行的信息。

两类记录为一对多关系：Session 的 `task_id` 关联 Task 的 `id`，一个 Task 可以包含多个 Session。
