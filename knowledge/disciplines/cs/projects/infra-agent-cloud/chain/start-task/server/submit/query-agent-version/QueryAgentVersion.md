# 查询 Agent 版本

读取 Agent 管理信息后，根据请求与版本选择规则，在 [AgentVersion](../../../../../storage/mysql/agent-version/AgentVersion.md) 中确定本次任务使用的版本记录。

## 1. 选择优先级

以下为分支优先级，命中后即返回，不会依次查询所有分支：

1. 请求使用草稿：选择草稿分支。
2. 请求指定版本记录 ID：按所属 Agent 和版本记录 ID 查询。
3. 请求指定版本号：按所属 Agent、版本号及已发布状态查询。
4. 未指定版本且启用多版本模式：尝试匹配灰度规则。
5. 未命中上述选择：查询默认发布版本。

## 2. 默认发布版本

默认发布版本需同时满足：

```text
AgentVersion.agent_id = Agent.id
AgentVersion.version  = Agent.published_version
AgentVersion.status   = 已发布
```

一个 Agent 可以保留多个版本，默认发布版本号仅指定通常使用的版本。`published_version` 对应版本号 `version`，不是版本记录的主键 `id`。

## 3. 查询失败与回退

显式指定的版本不存在或不属于当前 Agent 时，应返回错误，不能静默切换到默认版本。

无法找到默认发布版本时，只有允许草稿回退的调用才可尝试使用草稿，否则返回错误。草稿是否可执行还需结合权限与版本状态校验。

## 4. 查询结果

取得版本记录后，读取以下配置，交给后续配置准备流程：

- [模型](Model.md)：使用的模型及推理设置。
- [提示词](Prompt.md)：系统指令、角色与行为要求。
- [工具](Tools.md)：可调用的外部能力及连接配置。
- [技能](Skills.md)：完成特定任务的方法与配套资源。
- [插件](Plugins.md)：需要加载的扩展包及来源、版本信息。

下一步进入[组装执行配置](../AssembleExecutionConfig.md)，合并用户设置与本次请求配置，并检查授权。
