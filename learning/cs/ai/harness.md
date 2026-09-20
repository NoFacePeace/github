# Harness

Harness 是运行 Agent 的支架和执行环境。它位于模型与外部世界之间，负责驱动模型、连接工具、管理上下文，并控制任务的执行过程。

可以简单理解为：

- **模型**：负责理解和推理。
- **Harness**：负责提供工具、规则、状态和执行循环。
- **Agent**：在 Harness 中运行并完成任务的系统。

## 1. 核心职责

- 驱动模型与工具之间的多轮交互。
- 管理任务上下文、执行状态和中间结果。
- 提供文件系统、终端、浏览器、数据库或业务 API 等工具。
- 控制权限、审批、资源限制和安全边界。
- 处理超时、重试、暂停、恢复和停止。
- 记录执行轨迹，支持调试、评测和审计。

## 2. 与相关概念的区别

| 概念 | 主要职责 |
| --- | --- |
| 模型 | 理解输入、推理并生成文本或行动决策 |
| 工具 | 提供访问外部系统和执行具体操作的能力 |
| 工作流 | 按预先定义的步骤组织任务 |
| Agent 框架 | 提供构建 Agent 的抽象、组件和编程接口 |
| Harness | 在具体运行中组合模型、工具、状态、策略和执行循环 |

## 3. 编程 Agent 示例

一个编程 Agent 的 Harness 通常包含：

- 代码仓库和文件读写能力。
- 终端命令执行和测试环境。
- Git 操作、代码差异和提交管理。
- 上下文压缩、任务状态和检查点。
- 命令权限、人工确认和资源限制。
- 测试结果、错误信息和执行日志。

模型负责判断下一步应该阅读文件、修改代码、运行测试还是继续调查；Harness 负责执行这些动作，并将结果反馈给模型。

## 4. 开源框架、SDK 和 Harness

框架、SDK 和 Harness 都可以用于构建 Agent，但关注点不同：

| 概念 | 主要解决的问题 | 关注层次 |
| --- | --- | --- |
| 框架（Framework） | 如何组织和构建 Agent | 提供抽象、组件、生命周期和扩展机制 |
| SDK | 如何在代码中调用模型、平台或服务 | 提供 API、客户端、类型和工具函数 |
| Harness | Agent 如何真正运行并完成任务 | 负责工具执行、状态管理、权限、循环和运行环境 |

可以简单理解为：

- **框架**：Agent 的结构和组织方式。
- **SDK**：开发者可以调用的代码工具箱。
- **Harness**：让 Agent 实际运行起来的工作台和执行系统。

三者在实际项目中可能重叠。一个框架可以附带 SDK，Harness 也可以基于框架或 SDK 构建；同一个项目可能同时提供框架、SDK 和默认 Harness。

### 4.1 DeepSeek Harness

[DeepSeek Harness](https://github.com/deepseek-ai/deepseek-harness)（命令行工具：`dsh`）是 DeepSeek AI 开源的 Agent Harness。它既可以作为本地优先的 Coding Agent 使用，也可以作为开发和运行其他 Agent 的基础环境；目前处于开发者预览阶段。

#### 4.1.1 核心特点

- 基于 Cordis 构建，采用“一切皆插件”的架构。
- 模型适配器、工具注册、会话、Agent 循环、持久化和 UI 都可以通过插件组合。
- 内置或可扩展文件系统、Shell、Web、子 Agent、技能和会话等能力。
- 支持 Web、无界面运行方式，以及面向外部程序调用的 SDK。

#### 4.1.2 运行方式

```bash
npx @deepseek-ai/dsh web
```

如果需要在 Python 程序中驱动 Harness，可以使用 `deepseek-harness-sdk`，通过 JSON-RPC 与本地 Harness Runtime 通信。

DeepSeek Harness 适合作为研究 Agent 运行时、插件化架构、工具调用和本地 Coding Agent 的实践案例。由于项目仍在快速迭代，使用时应关注版本兼容性和安全说明。

### 4.2 Vercel AI SDK

[Vercel AI SDK](https://ai-sdk.dev/) 是 Vercel 提供的开源 TypeScript SDK，用于在 JavaScript/TypeScript 应用中接入大语言模型并构建 AI 功能。它支持 Next.js、React、Vue、Svelte、Node.js 等环境，并统一不同模型供应商的调用方式。

#### 4.2.1 主要能力

- 文本、结构化数据和流式输出。
- Chat UI 和生成式 UI。
- 工具定义、工具调用和人工审批。
- Agent 循环、状态管理和工作流。
- 不同模型供应商和模型接口的适配。

#### 4.2.2 与 Harness 的区别

Vercel AI SDK 主要解决“如何在代码中调用模型并构建 AI 应用”的问题，属于 SDK；Harness 主要解决“如何让 Agent 在具体环境中持续运行并完成任务”的问题，负责工具执行、权限、状态和运行循环。

Vercel AI SDK 可以用来构建 Agent，也可以作为 Harness 中的模型调用和工具编排组件，但它本身不等同于完整的 Harness。
