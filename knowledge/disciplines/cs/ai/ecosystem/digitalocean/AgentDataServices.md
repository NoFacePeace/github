# DigitalOcean Agent Data Services

Agent Data Services 是 DigitalOcean 为 Agent 提供的托管数据服务。可以把它理解成一个“资料室”：将文档、多媒体、结构化记录和 Agent 执行轨迹整理为可查询的数据，供 Agent 获取上下文，也供开发者分析运行情况。

## 1. 核心能力

- 多模态数据处理：提取文档内容，对音视频转录、描述，并进行分块和向量化。
- 行为数据处理：将日志、事件和 Agent 轨迹组装为会话，并按配置对敏感信息脱敏。
- 统一数据目录：保存处理结果及其来源，便于追踪数据经过哪些转换。
- 统一查询：支持语义与关键词搜索、按标识查询完整会话，以及过滤和聚合。

## 2. 与检索、记忆和会话的关系

| 方面 | 对应关系 |
|---|---|
| [知识库与检索](../../agent/retrieval/Retrieval.md) | 整理外部资料、建立索引，供 Agent 查找相关信息 |
| [记忆](../../agent/memory/Memory.md) | 保存并检索会话和执行记录，可作为记忆机制的数据基础 |
| [会话](../../agent/session/Session.md) | 将执行事件组装为会话，支持按会话标识查询历史交互 |
| 可观测性与评估 | 为运行分析和测试提供历史轨迹与上下文 |

存储历史记录并不自动构成完整的记忆系统。哪些信息需要记住、何时召回、如何更新或遗忘，仍需由 Agent 应用设计。

## 3. 接入方式

这是托管云服务，官方文档主要介绍控制台和 API 接入。基本流程如下：

1. 申请并获得私有预览访问权限。
2. 创建 Workspace（工作区），配置访问权限。
3. 定义 Data Flow（数据处理流程），组合提取、分块、向量化、脱敏或会话组装等内置功能。
4. 通过支持的路径导入资料或执行事件，由数据流程处理并建立索引。
5. 在控制台或通过查询 API 检索数据，将结果提供给 Agent。

API 中仍使用旧名称：`sandbox` 对应工作区，`pipeline` 对应数据处理流程。应以官方 API 定义为准，避免直接用产品界面名称猜测接口。

## 4. 开放状态与限制

截至 2026 年 10 月 19 日，Agent Data Services 处于私有预览，仅向获邀团队开放：

- 支持控制台和 HTTP API，暂不支持 `doctl`、官方客户端库或 MCP Server。
- 查询接口为只读；数据写入通过导入和数据处理流程完成。
- 数据处理流程使用内置函数，暂不支持客户自定义函数。
- 底层表和向量数据库不直接开放访问，也不能接入外部向量数据库。
- 暂无 CRM、工单系统等第三方系统连接器；数据库中的数据需先导出再导入。

预览期间的功能与限制可能变化，使用前应核对官方文档。

## 5. 参考资料

- [产品介绍](https://www.digitalocean.com/blog/introducing-agent-data-services)
- [官方文档](https://docs.digitalocean.com/products/agent-data-services/)
- [运行原理](https://docs.digitalocean.com/products/agent-data-services/concepts/how-agent-data-services-works/)
- [API 接入说明](https://docs.digitalocean.com/products/agent-data-services/reference/)
- [访问控制](https://docs.digitalocean.com/products/agent-data-services/details/access-control/)
- [预览限制](https://docs.digitalocean.com/products/agent-data-services/details/limits/)
