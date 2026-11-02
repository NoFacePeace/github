# Claude Agent SDK

Claude Agent SDK 是 Anthropic 提供的智能体开发工具包，用于将 Claude Code 的工具使用与多步任务执行能力集成到自己的应用中。开发者提交目标、配置工具和权限，并接收执行过程与结果。本文依据官方 Python 仓库整理，核对日期为 2026-11-01；源码与开源范围的说明以 Python SDK 为准。

## 1. 产品定位

| 组成部分 | 作用 |
|---|---|
| Claude 模型 API | 提供模型推理与工具调用请求等基础能力 |
| Claude Code 运行时 | 组织模型与工具交替执行，处理内置工具、上下文和任务运行流程 |
| Claude Agent SDK | 提供编程接口，让应用启动、配置、扩展和控制上述执行过程 |

例如，应用提交“修复项目中的一个 bug”后，Agent 可以读取文件、修改代码、运行测试，再根据结果继续调整。SDK 适合构建代码助手、文件处理助手和运维自动化等需要多步工具操作的应用；简单问答通常直接调用模型 API 即可。

## 2. 执行流程与 Agent Loop

完整的消息生命周期与轮次限制见 [Agent Loop](AgentLoop.md)。

Agent Loop 指围绕目标反复执行以下步骤的循环：

```text
请求模型 → 模型提出工具调用 → 执行工具 → 将结果交回模型
   ↑                                          ↓
   └──────────── 继续下一轮，直到完成或达到限制 ──────┘
```

Claude Agent SDK 提供的整体能力包含这个循环，调用方无需自行编写核心 Agent Loop。但在默认的 Python SDK 实现中，核心循环由底层 Claude Code CLI 运行时承担，SDK 通过子进程通信接入。

```text
你的应用
   ↓ 提交任务、配置与回调
Python SDK
   ↓ 启动 CLI，通过消息流和控制协议通信
Claude Code CLI 运行时
   ↔ Claude 模型服务
   ↔ 内置工具、外部 MCP 工具、SDK 自定义工具
```

SDK 示例中的 `async for message in query(...)` 用来消费消息流，不是驱动模型与工具交替执行的核心循环。SDK 自身也有消息读取、事件分发和回调处理循环，需要与 Agent Loop 区分。

## 3. 公开源码的内容

Python SDK 的公开源码主要实现应用与运行时之间的接入、通信和扩展能力。

| 内容 | 具体职责 | 代表源码 |
|---|---|---|
| 调用接口 | 提交任务、双向交互、接收响应 | `query.py`、`client.py` |
| 进程管理与通信 | 查找并启动 CLI、传递参数、读写消息流、清理进程 | `_internal/transport/subprocess_cli.py` |
| 消息与类型 | 解析消息，定义配置、内容块和错误类型 | `_internal/message_parser.py`、`types.py`、`_errors.py` |
| 控制协议与回调 | 处理权限请求、Hooks、执行中断等交互 | `_internal/query.py` |
| 自定义工具接入 | 将 Python 函数注册为进程内 MCP 工具，处理调用和结果 | `tool()`、`create_sdk_mcp_server()` 及 MCP 桥接实现 |
| 配套工程 | 示例、测试、构建与 CLI 打包 | `examples/`、`tests/`、`scripts/` |

开发者可以通过这些代码了解接口设计、消息通信、权限回调和工具扩展方式。Claude Code 的核心 Agent Loop、内置工具执行和上下文管理的完整实现，不在该 Python SDK 开源源码中。

## 4. 安装包与开源边界

官方 Python 安装包默认捆绑 Claude Code CLI，无需单独安装；也可以通过 `ClaudeAgentOptions(cli_path="/path/to/claude")` 指定其他 CLI 路径。

因此需要区分：

- **运行能力包含核心循环**：安装后可以直接使用完整的多步任务执行能力。
- **源码不包含核心循环的完整实现**：捆绑 CLI 可执行程序，不等于公开其源码。
- **SDK 源码许可与依赖许可分别判断**：Python 仓库的 `LICENSE` 为 MIT；捆绑 CLI、模型服务及具体使用还需查看各自适用的许可和服务条款，不能将仓库的 MIT 许可直接套用到所有组件。

仅凭 Python SDK 的公开源码，无法重建整套 Claude Code 运行时，也不意味着可以离线自托管 Claude 模型。

## 5. 用户交互

- [用户输入与审批](UserInput.md)：通过权限回调处理工具审批与澄清问题，在任务执行期间等待用户决策后继续。

## 6. 会话管理

- [会话](Sessions.md)：保存对话历史，通过 Continue、Resume 和 Fork 继续或分叉会话，以及跨主机恢复的条件。

## 7. 参考资料

- [Claude Agent SDK 官方文档](https://platform.claude.com/docs/en/agent-sdk/overview)
- [Python SDK 仓库与 README](https://github.com/anthropics/claude-agent-sdk-python)
- [Python SDK MIT 许可证](https://github.com/anthropics/claude-agent-sdk-python/blob/main/LICENSE)
- [CLI 子进程通信实现](https://github.com/anthropics/claude-agent-sdk-python/blob/main/src/claude_agent_sdk/_internal/transport/subprocess_cli.py)
- [双向控制协议实现](https://github.com/anthropics/claude-agent-sdk-python/blob/main/src/claude_agent_sdk/_internal/query.py)
