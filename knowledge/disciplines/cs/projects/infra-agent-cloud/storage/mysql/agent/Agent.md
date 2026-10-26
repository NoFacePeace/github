# Agent

Agent 表保存智能体的身份和管理信息。名称、描述以及模型、提示词等执行配置由版本记录保存。

## 1. 相关字段

| 字段 | 含义 |
| --- | --- |
| [id](ID.md) | Agent 主键。 |
| status | 启用状态。 |
| visibility | 可见范围。 |
| [published_version](PublishedVersion.md) | 当前发布版本号。 |
| has_draft | 是否存在草稿。 |
| creator | 创建者。 |

## 2. 关联关系

版本记录通过 `agent_id` 关联 Agent 的 `id`；任务也可通过 `agent_id` 记录所使用的 Agent。
