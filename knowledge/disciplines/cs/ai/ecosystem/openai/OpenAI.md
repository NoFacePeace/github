# OpenAI

OpenAI 提供模型、面向用户的 AI 产品，以及模型 API 和 Agent 开发能力。本目录按厂商整理其产品与平台。

## 1. 产品

- [ChatGPT](ChatGPT.md)：面向通用对话与工作任务的 AI 助手产品，包含对话与多步任务执行等使用方式。
- [Codex](codex/Codex.md)：面向软件开发的 Agent 产品与运行环境，提供项目操作及开发者视图，也能处理其他知识工作。

## 2. 产品与开发接口

ChatGPT、Codex 是面向用户的产品；模型 API、Agents SDK 和 Agents API 是开发者构建应用或运行 Agent 的接口与框架。产品、模型和运行框架属于不同层次。

官方 Agents API 使用托管的 Codex harness；Agents SDK 则在开发者应用中运行 Agent 循环。使用 OpenAI 构建 Agent 不要求使用 Codex 客户端。

## 3. 参考资料

- [ChatGPT 与 Codex 的使用方式](https://learn.chatgpt.com/docs/use-chatgpt)
- [Agent 开发路径对比](https://developers.openai.com/api/docs/guides/agents)
- [插件架构](https://developers.openai.com/plugins/concepts/plugins)
