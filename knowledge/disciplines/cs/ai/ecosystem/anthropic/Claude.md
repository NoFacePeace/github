# Claude

Claude 是 Anthropic 的模型与产品品牌。本文说明其模型、应用、Agent 产品与托管平台的定位关系。

## 1. 产品与平台定位

| 名称 | 定位 |
|---|---|
| Claude 模型 | 提供推理、内容生成与工具调用等能力；Agent 系统可围绕模型构建执行循环 |
| Claude 应用 | 面向用户的工作助手，支持问答及工具驱动的任务执行 |
| Claude Code | Agentic coding tool，能够读取代码、编辑文件、运行命令的编程工具 |
| [Claude Managed Agents](managed-agents/ManagedAgents.md) | 预构建、可配置的 Agent Harness 与托管运行平台，适合长时间任务和异步工作 |

模型访问与 Agent 运行平台属于不同层次。Messages API 提供直接模型访问，开发者可自行实现 Agent Loop；Managed Agents 提供执行循环、工具执行和运行环境等能力。

## 2. 参考资料

- [Claude 产品介绍](https://claude.com/product/overview)
- [Claude Code 概览](https://code.claude.com/docs/en/overview)
- [Claude Managed Agents 概览](https://platform.claude.com/docs/en/managed-agents/overview)
- [多 Agent 编排](https://platform.claude.com/docs/en/managed-agents/multiagent-orchestration)

[返回 Anthropic](Anthropic.md)
