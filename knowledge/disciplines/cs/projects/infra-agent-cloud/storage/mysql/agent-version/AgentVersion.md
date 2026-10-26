# AgentVersion

AgentVersion 表保存智能体各版本的展示信息与执行配置。同一 Agent 可以拥有多个版本。

## 1. 相关字段

| 字段 | 含义 |
| --- | --- |
| [id](ID.md) | 版本记录主键。 |
| [agent_id](AgentID.md) | 所属 Agent 的主键。 |
| [version](Version.md) | 版本号，与记录主键不同。 |
| status | 草稿、发布等版本状态。 |
| name / description | 该版本的名称与描述。 |
| engine / model | 执行引擎与模型配置。 |
| system_prompt | 系统提示词。 |
| mcp_servers | MCP 工具配置。 |
| skills_config | 技能配置。 |
| agent_plugins_config | 插件配置。 |

## 2. 关联关系

`agent_id` 指向 Agent 的 `id`。任务通过 `version_id` 关联本表的 `id`，并非关联 `version` 版本号。
