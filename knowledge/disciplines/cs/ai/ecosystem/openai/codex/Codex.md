# Codex

Codex 是 OpenAI 面向软件开发的 Agent 产品与运行环境，可围绕项目读取和修改文件、执行命令、运行验证并审阅变更。它也能完成研究、文档和演示文稿等知识工作。

## 1. 工作方式

典型流程是理解任务、检查项目、实施修改、运行适当验证并交付可审阅的结果。实际操作能力由所连接的环境、工具与权限决定。

Codex 提供 CLI、IDE 和桌面等使用入口；不同入口的界面与可用能力应以各自文档和版本为准。

## 2. 与 ChatGPT 的区别

[ChatGPT](../ChatGPT.md) 与 Codex 都可以采用 Agent 的工作方式。Codex 更强调项目和开发者操作，桌面界面可展示 Git、Shell、代码差异与审阅等细节；ChatGPT 更强调通用任务与面向结果的表达。

这一差别属于产品界面、工具和工作流程的定位，不表示两者使用完全不同的能力体系。产品说明核对日期为 2026-10-20。

## 3. 扩展与执行控制

- Skill：提供任务说明和配套资源，帮助 Agent 使用已有工具。
- 工具与 MCP：接入文件、终端和外部服务等操作能力。
- [插件](Plugins.md)：打包、安装和分发技能、工具连接及其他支持的扩展。
- [Hook](Hooks.md)：由宿主在生命周期事件发生时触发操作，可独立配置或随插件分发。

配置入口和支持范围取决于版本。普通用户、项目和插件提供的非托管 Hook 需要审阅和信任；通用概念见 [Agent Hooks](../../../agent/hooks/Hooks.md)。

## 4. 与 Codex Harness 的关系

Harness 是围绕模型提供执行循环、工具调度和上下文管理等能力的运行框架。官方 Agents API 提供托管的 Codex harness，开发者可以通过 API 使用相关运行能力，无需通过 Codex 客户端发起任务。

Codex 产品与 Codex harness 需要区分：前者是用户使用的产品，后者是运行 Agent 的框架。

## 5. 参考资料

- [ChatGPT Work 与 Codex 的桌面差异](https://learn.chatgpt.com/docs/use-chatgpt)
- [Codex 使用与开发资源](https://learn.chatgpt.com/learn/codex)
- [Codex Hooks](https://learn.chatgpt.com/docs/hooks)
- [Agent 运行方式对比](https://developers.openai.com/api/docs/guides/agents)
