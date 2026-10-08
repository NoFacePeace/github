# AI 与 Agent 新闻摘要

整理日期：2026 年 10 月 8 日。检索范围：北京时间 2026 年 10 月 7 日至 10 月 8 日晚 20:30 左右。

本文整理本次检索中核对过的官方公告，不代表全部新闻。以下日期保留原站标注；只有日期的来源不作精确时区换算。厂商性能、成本与产品能力描述按公告理解，不能直接推广为所有业务场景的结果。

## 1. Anthropic 发布 Claude Haiku 5.5

官方公告日期：2026 年 10 月 7 日。

Claude Haiku 5.5 面向摘要、分类、数据库查询和编程子 Agent 等高频、成本敏感任务，新增可调节的推理 effort 设置。Anthropic 表示该模型已在 Claude Platform、AWS、Google Cloud 和 Microsoft Azure 提供，模型 ID 为 `claude-haiku-5-5`。

### 1.1 价格与配套调整

公告列出的每百万 token 价格如下：

| 项目 | 提示长度不超过 10 万 token | 提示长度超过 10 万 token |
|---|---:|---:|
| 输入 | 0.10 美元 | 0.50 美元 |
| 输出 | 0.50 美元 | 2.50 美元 |
| 缓存读取 | 0.01 美元 | 0.05 美元 |
| 缓存写入 | 0.125 美元 | 0.625 美元 |

同日，Sonnet 5.5 的缓存读取价格从每百万 token 0.20 美元降至 0.10 美元。Anthropic 估计这会使多数 Agent 任务的运行成本降低约 20%；实际收益取决于缓存命中与 token 构成。

### 1.2 对 Agent 开发的意义

可以尝试由大模型负责规划，将范围明确的摘要、分类、查询和信息提取交给小模型。复杂编程任务仍需结合实际质量与成本评测选择模型，公告也指出 Sonnet 5.5 和 Opus 5.5 更适合复杂 Agent 编程工作。

来源：[Anthropic 官方公告](https://www.anthropic.com/claude-haiku-5-5)。

## 2. 微软正式推出 Agent 执行隔离层 MXC

官方公告日期：2026 年 10 月 7 日。

Microsoft Execution Containers（MXC）正式可用。开发者和管理员可以定义 Agent 允许访问的文件、网络与桌面资源，由运行环境在执行时强制落实。策略位于 Agent 工作负载的控制范围之外，Agent 或生成的代码不能自行扩展权限。

### 2.1 隔离方式与上线状态

| 后端 | 公告中的支持范围 | 用途与状态 |
|---|---|---|
| 进程容器 | Windows 11、macOS、Linux | 轻量执行隔离，适合工具和生成代码 |
| 会话容器 | Windows 11 | 独立账户与会话，分离桌面、剪贴板和输入 |
| WSL 容器 | Windows 11 | 依赖 Linux 工具链的工作负载 |
| MicroVM | Windows 11、Linux | 硬件虚拟化隔离，仍属实验性 |

公告列出 Enforcement、Learning 和 Permissive 三种模式。Learning 会阻止并记录未授权访问；Permissive 会记录但允许策略原本会拒绝的访问，适合观察和编写策略，不能将其理解为已强制执行该策略。

### 2.2 对工程实践的意义

将权限边界落实在执行层。例如，允许编程 Agent 修改代码仓库、读取部署配置，同时禁止它修改生产配置或访问无关个人文件。

来源：[微软官方说明](https://blogs.windows.com/windowsdeveloper/2026/10/07/microsoft-execution-containers-policy-driven-containment-for-ai-agents/)。

## 3. Windows 推进本地与云端混合 Agent

官方公告日期：2026 年 10 月 7 日。

微软宣布将 GitHub HydraFusion 路由扩展到 Windows 本地模型，让任务根据需要在本地与云端执行。该能力计划在 10 月稍后向 GitHub Copilot app、GitHub Copilot CLI 和 Visual Studio Code 提供实验预览。

微软还宣布 Copilot 将结合本地文件与近期活动上下文，执行文件整理、诊断、排障和编程等本地操作，并在合适的任务上使用本地模型。这些混合智能功能预计未来几个月开始在 Copilot+ PC 上逐步推出，不能理解为公告当天已经全部可用。

对 Agent 开发而言，值得观察的是本地与云端任务分配如何影响延迟、成本、数据访问和权限控制。

来源：[微软官方公告](https://blogs.windows.com/windowsexperience/2026/10/07/building-windows-for-hybrid-intelligence/)。

## 4. LlamaIndex 发布 OpenDocRouter

官方公告日期：2026 年 10 月 7 日。

OpenDocRouter 用统一 API 接入多种开源和商业文档解析模型，将 PDF、PNG、JPEG 或对应文件 URL 转换为 Markdown。每个模型使用带版本的解析配置，包含提示、处理流程和设置。

### 4.1 主要能力

- 支持同步与异步请求；同步最多支持 50 页，输入上限为 500 页或 50 MB。
- 可启用版面解析，返回阅读顺序、版面类别和内容边界框。
- 便于切换模型、比较质量和成本，减少分别部署与适配的工作。
- 按 token 计费，开启版面解析会产生额外费用。

### 4.2 对 RAG 与文档 Agent 的意义

可作为知识库和文档 Agent 的解析入口，先将文档转为带结构与位置的文本，再交给切分、检索或提取流程。它提供文档解析能力，完整 RAG 系统仍需检索、权限过滤、生成和评估等组件。

来源：[LlamaIndex 官方公告](https://www.llamaindex.ai/blog/introducing-opendocrouter)。

## 5. Apollo 发布 GraphOS Agent Services

官方公告日期：2026 年 10 月 7 日。

GraphOS Agent Services 位于 Agent 与企业系统/API 之间，提供工具和数据发现、身份与凭证管理、字段级权限控制以及审计。公告称 Intuit 正在预览试点。

其策略层控制人或 Agent 能看到和执行的内容，并在字段级落实规则，权限判断不交给语言模型。Apollo 同时扩展 GraphOS MCP Server，增加构建和管理 graph 的工具。

对企业 Agent 的意义：调用已有业务 API 时，可以限制返回字段、操作权限与凭证范围，并留下可追溯记录。例如，查询客户姓名不应自动暴露无关账单或内部备注。

来源：[Apollo 官方公告](https://www.apollographql.com/newsroom/press-releases/apollo-graphql-introduces-graphos-agent-services)。

## 6. Sierra 发布 fleming-1，识别打电话的 AI Agent

官方公告日期：2026 年 10 月 8 日。

fleming-1 实时分析来电语音，判断其中是否存在 AI 生成音频的迹象，并标记可能由 AI 发起的来电。企业自行决定后续处理，例如增加验证步骤或统计自动来电比例。

Sierra 表示该能力可用于其平台上的语音 Agent。公告未提供完整准确率与误判率数据，因此不能据此判断其识别可靠性，也不能把“可能为 AI 来电”直接等同于欺诈。

这反映了一个新的应用场景：用户让个人 Agent 代打电话，企业客服需要识别来电主体，并设计对应的授权和交互流程。

来源：[Sierra 官方文章](https://sierra.ai/blog/caller-id-in-the-age-of-agents)。

## 7. 阅读优先级

结合 Agent 工程学习方向，优先阅读以下三项：

| 新闻 | 关注问题 |
|---|---|
| Haiku 5.5 | 如何分配大小模型任务并控制成本 |
| MXC | 如何在执行层强制落实 Agent 权限边界 |
| GraphOS Agent Services | 如何管理 Agent 对企业 API 的身份、字段权限与审计 |

本次摘要排除了部分在 10 月 7 日新闻聚合页面出现、但官方公告日期为 10 月 6 日的内容，避免把重复报道当作新发布。
