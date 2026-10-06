# AI 与 Agent 新闻摘要

覆盖时间：北京时间 2026 年 10 月 5 日至 10 月 6 日截至整理时。整理日期：2026 年 10 月 6 日。来源日期保留原站标注；只标日期的公告无法精确换算到北京时间。

本期重点是编程 Agent 更新、Skill 工具按需加载、模型接入和业务 Agent 的持续改进。厂商性能与成本数据属于官方表述，未经本文独立验证；案例发布日期也不等于产品首次上线日期。

## 1. Agent 工具与框架

### 1.1 Claude Code 更新子 Agent、后台任务与权限处理

Claude Code v2.1.290 的官方发布记录标注 10 月 5 日，新增插件权限检查事件中的 `agentId`，方便区分子 Agent 与主会话。它还修复了子 Agent 恢复时丢失思考记录和提示缓存、定时任务恢复异常，以及 WebFetch 静默截断长页面等问题。

10 月 6 日发布的 v2.1.291 进一步修复云会话丢失权限提示回答，以及退出时丢失最后几条消息的回归问题。

来源：[v2.1.290 发布说明](https://github.com/anthropics/claude-code/releases/tag/v2.1.290)、[v2.1.291 发布说明](https://github.com/anthropics/claude-code/releases/tag/v2.1.291)。

### 1.2 Deep Agents 0.7.22 支持 Skill 工具按需加载

LangChain 于 10 月 5 日发布 Deep Agents 0.7.22，增加读取 Skill 后才向模型提供其关联工具的能力。

Skill 可以在 `metadata.include_tools` 中声明工具名称，开发者通过 `SkillsMiddleware` 注册工具或解析函数。只有 Skill 的读取记录仍在对话中，对应工具才可调用；上下文压缩移除记录后，需要重新读取。

应用判断：适合工具很多的 Agent，可将操作指南与执行能力一起按需提供，减少无关工具进入上下文。工具可见性仍需配合实际权限校验。

来源：[版本发布说明](https://github.com/langchain-ai/deepagents/releases/tag/deepagents%3D%3D0.7.22)、[实现与使用说明 PR #6552](https://github.com/langchain-ai/deepagents/pull/6552)。

### 1.3 Together Link 连接编程 Agent 与开放模型

Together AI 于 10 月 5 日发布 Together Link，可将 Claude Code、Codex、OpenCode 等接入 Together AI 上的开放模型，并提供按会话自动路由和费用跟踪。

官方宣称可降低超过 50% 的模型支出。实际收益需要结合任务完成质量、重试次数和模型价格评估。

来源：[Together AI 官方公告](https://www.together.ai/blog/together-link-frontier-quality-open-models-in-the-harness-you-already-use)。

## 2. 业务 Agent 实践

### 2.1 Cresta 分享基于 Claude Agent SDK 的 Conductor 案例

Anthropic 于 10 月 5 日发布 Cresta 案例。Conductor 帮助团队构建和改进客服 Agent：Claude Agent SDK 负责上下文收集、工具调用和代码执行，Cresta 在上层加入客服业务知识、流程控制与评测。

其流程包括构建、评测与反馈改进，并将有效经验保存为可复用的 Skills。评测关注结果、执行路径、质量和资源消耗。这是实践案例发布，不能视为当天首次推出产品。

应用判断：长期维护业务 Agent 时，应让失败案例成为后续评测输入，并在更换模型、提示词或框架后重新验证。

来源：[Anthropic 官方案例](https://claude.com/blog/how-cresta-turned-cx-expertise-into-an-agent-builder-on-the-claude-agent-sdk)。

## 3. Agent 模型与部署

### 3.1 Reflection 公布首个开放权重模型 Beam

Reflection 于 10 月 5 日公布 Beam，采用 MoE 架构，总参数 501B、激活参数 23B，面向编程、推理和 Agent 任务。

目前提供提前体验，权重与技术报告计划本月晚些时候发布，权重计划采用 Apache 2.0 许可。公告中的能力与效率比较属于厂商评测。

来源：[Reflection 官方公告](https://reflection.ai/blog/introducing-beam)。

### 3.2 GLM 5.3 上线 Amazon Bedrock

AWS 于 10 月 5 日宣布向符合条件的企业客户提供 GLM 5.3 托管访问。模型面向编程和长流程 Agent 任务，支持 OpenAI 兼容 API、提示缓存、跨区域推理和服务层级。

此次消息是 Bedrock 接入更新，不代表 GLM 5.3 在当天首次发布。

来源：[AWS 官方公告](https://aws.amazon.com/blogs/machine-learning/introducing-glm-5-3-on-amazon-bedrock/)。

## 4. Agent 安全与监管

### 4.1 维基媒体披露未经授权的 Agent 活动

路透社于 10 月 5 日报道，维基媒体基金会确认发现 OpenAI Agent 在其平台上的未经授权活动，并表示大量访问与查询可能与今年 5 月 Wikidata 查询服务的部分中断有关。OpenAI 表示正与其合作分析。

这是新的调查披露，故障发生在 5 月，因果关系尚未确定。

来源：[路透社报道，由 KFGO 转载](https://kfgo.com/2026/10/05/wikipedia-operator-says-openais-rogue-agents-possibly-tied-to-data-service-disruption-in-may/)。

### 4.2 Anthropic 对强制披露 AI 入侵事件的立法持开放态度

据路透社 10 月 6 日报道，Anthropic 在澳大利亚议会听证会上表示，对要求 AI 公司披露数据泄露事件的法律持开放态度。这是公司政策立场，不代表相关法律已经通过。

来源：[路透社报道，由 InnovationAus 转载](https://www.innovationaus.com/anthropic-open-to-aussie-laws-requiring-reporting-of-ai-hacks/)（本次摘要依据检索结果，全文未能访问）。

## 5. 开发者关注重点

以下为基于本期新闻的应用判断：

- **工具按需提供。** Deep Agents 展示了将 Skill 指南与关联工具一起加载的机制，值得用于工具规模较大的 Agent。
- **建立持续评测流程。** Cresta 案例说明，业务 Agent 需要围绕真实任务、失败反馈和经验复用持续迭代。
- **按任务比较模型成本。** Together Link、Beam 和 Bedrock 的更新提供更多接入选择，评估时应看任务总成本和完成质量。
- **验证长任务与权限行为。** Claude Code 的修复及安全披露提示开发者关注恢复、上下文压缩、权限检查与行动记录。

建议优先阅读 Deep Agents 的实现说明和 Cresta 案例，分别研究运行时能力管理与业务 Agent 的开发改进流程。
