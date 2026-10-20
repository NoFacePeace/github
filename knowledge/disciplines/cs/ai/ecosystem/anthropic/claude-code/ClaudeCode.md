# Claude Code

Claude Code 是 Anthropic 的编程 Agent，能够读取代码库、编辑文件、运行命令并集成开发工具，用于功能开发、问题修复和开发流程自动化。本文依据官方文档整理，核对日期为 2026-10-20。

## 1. 产品定位

| 名称 | 定位 |
|---|---|
| [Claude](../Claude.md) | 模型与产品品牌 |
| Claude Code | 面向代码库与开发任务的 Agent 产品及执行环境 |
| Claude Managed Agents | 托管 Agent 运行平台 |

Agent 是工作方式的通用概念，不是 Claude 或 Claude Code 的专有名称。Claude Code 围绕目标检查项目、使用工具、观察执行结果并继续推进任务。

## 2. 使用入口与工作流程

官方提供终端、IDE、桌面和网页等入口，各入口的运行位置、账号要求与功能范围应以对应文档为准。

典型流程：理解任务、检查代码、制定修改方案、编辑相关文件、运行验证并审阅差异。可用于跨文件修改、测试、代码审查及 Git 工作流；实际执行受工具、环境和权限配置约束。

## 3. 扩展机制

| 机制 | 用途 |
|---|---|
| `CLAUDE.md` | 项目约定、架构说明及工作规则 |
| Skills | 可复用的任务说明、脚本与参考资源 |
| Subagents | 将专门任务交给独立上下文中的子 Agent |
| MCP | 连接外部工具和数据服务 |
| [Hooks](Hooks.md) | 在生命周期事件发生时触发操作 |
| [Plugins](Plugins.md) | 将多种扩展打包、安装和分发 |

Hook 可独立配置，也可随插件分发。脚本可以由已有工具执行，不会仅因存在于目录中就自动成为独立工具。

## 4. 官方文档

- [Claude Code 概览](https://code.claude.com/docs/en/overview)
- [工作原理](https://code.claude.com/docs/en/how-claude-code-works)
- [扩展机制](https://code.claude.com/docs/en/features-overview)
- [插件清单参考](https://code.claude.com/docs/en/plugins-reference)
