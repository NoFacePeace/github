# Claude Managed Agents

Claude Managed Agents 是 Anthropic 提供的预构建、可配置的 Agent Harness 与托管运行平台，适合长时间任务和异步工作。开发者通过 API 配置 Agent、提交任务并接收进度与结果，平台负责执行循环、工具执行和会话运行管理。

## 1. 与模型 API 和应用的区别

| 入口 | 使用方式 | 执行流程由谁管理 |
|---|---|---|
| Messages API | 程序直接调用模型 | 开发者自行实现 Agent Loop、工具执行和状态管理 |
| Managed Agents | 程序通过 API 创建会话并发送任务 | 平台提供 Agent Harness 与运行管理 |
| Claude 应用 | 用户通过产品界面提交任务 | 由应用提供交互体验与任务执行能力 |

Managed Agents 面向开发者集成自己的系统；Claude 应用面向用户直接使用。功能相似不意味着应用底层一定调用公开的 Managed Agents API。

## 2. 核心对象

| 对象 | 含义 |
|---|---|
| Agent | 模型、系统提示词、工具、MCP 服务和 Skills 等配置 |
| Environment | 会话的执行环境配置，可使用托管云沙箱或自托管沙箱 |
| Session | 引用 Agent 和 Environment 的有状态任务会话 |
| Events | 用户消息、工具结果、状态更新等交互事件 |

Agent 是可复用的配置；Session 是实际执行任务的实例。

## 3. 使用流程

1. 定义并创建 Agent，获得其标识。
2. 创建 Environment，配置执行环境。
3. 创建 Session，引用 Agent 和 Environment。
4. 发送用户消息，通过 SSE 事件流读取执行输出和状态。
5. 根据需要发送补充要求、处理自定义工具请求或中断任务。

平台支持文件操作、Shell 命令、网页检索与 MCP 等工具。托管平台并不意味着所有自定义工具都由服务端实现；客户端仍需处理配置为自定义工具的调用。

## 4. 多 Agent 能力

- Subagent：主 Agent 委派任务，读取结果后决定下一步，并可继续沟通。
- [动态工作流](DynamicWorkflows.md)：主 Agent 生成编排程序，服务端后台执行多个 Agent、传递并汇总结果。
- Advisor：主 Agent 咨询其他模型，具体工作仍由主 Agent 执行。

## 5. 接入与状态管理

Managed Agents 当前处于 Beta，请求需要 Claude API Key 与 `managed-agents-2026-04-01` Beta Header，官方 SDK 自动设置该 Header。

会话在服务端保存事件历史和运行状态，可以读取历史事件。开发者应保存会话标识，并按业务需求配置权限、预算和数据清理策略。功能与接口以具体版本的官方文档为准。

## 6. 参考资料

- [Managed Agents 概览](https://platform.claude.com/docs/en/managed-agents/overview)
- [多 Agent 编排](https://platform.claude.com/docs/en/managed-agents/multiagent-orchestration)
- [Workflow Runs](https://platform.claude.com/docs/en/managed-agents/workflow-runs)

[返回 Claude](../Claude.md)
