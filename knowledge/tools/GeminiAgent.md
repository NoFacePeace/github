---
color: "#4A90D9"
---

# Gemini Agent

Gemini Agent 是 Google 面向企业工作的 Agent 产品，用于处理问答、知识工作、内容生成和代码执行，并通过工具连接业务系统。

信息核对日期：2026 年 10 月 9 日。以下依据 Gemini at Work 2026 官方公告整理；功能开放范围以具体账户和官方产品文档为准。

## 1. 主要能力

| 能力 | 官方公告中的说明 |
| --- | --- |
| 持续执行 | 在云端执行任务，支持跨会话记忆和长时间工作 |
| 多 Agent 协作 | 创建子 Agent，协调并行或顺序执行的步骤 |
| 系统连接 | 连接 Workspace、Slack、Git、Jira 等工具及 MCP 服务 |
| 模型选择 | 根据任务选择 Gemini 或 Claude 模型 |
| 企业治理 | 提供身份、授权、审计、执行隔离和成本控制 |

## 2. 开源状态

截至核对日期，未找到完整 Gemini Agent 产品的源码仓库或开源许可证。官方将其作为企业云产品介绍。

Google 的 [Agent Development Kit（ADK）](../disciplines/cs/ai/agent/frameworks/GoogleADK.md) 是开源 Agent 开发框架，可以用于自行构建 Agent；ADK 开源不代表 Gemini Agent 产品开源。

## 3. 相关知识

- [AI Agent](../disciplines/cs/ai/agent/AIAgent.md)：智能体通用知识入口。

## 4. 参考资料

- [Gemini at Work 2026 官方公告](https://cloud.google.com/blog/products/ai-machine-learning/welcome-to-gemini-at-work-2026)。
- [ADK 官方介绍](https://developers.googleblog.com/en/agent-development-kit-easy-to-build-multi-agent-applications/)。
