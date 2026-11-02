# Codex Harness

Codex Harness 是 OpenAI 在 [openai/codex](https://github.com/openai/codex) 仓库中开源的 Agent 运行框架及相关集成组件。该仓库包含 Codex CLI、核心 Agent Loop、app-server 与 SDK，采用 Apache 2.0 协议，核心实现主要使用 Rust。本文记录开源项目本身，核对日期为 2026-11-02。

## 1. 仓库与组件

| 路径 | 内容 |
|---|---|
| `codex-rs/` | Rust 实现的 CLI、运行框架及配套组件 |
| `codex-rs/core/` | Agent 核心运行能力 |
| `codex-rs/app-server/` | 面向客户端集成的 app-server |
| `sdk/` | 程序化集成 SDK |
| `docs/` | 项目文档 |

源码布局随版本变化，阅读时以目标版本为准。仓库提供的不只是终端界面，还包括驱动模型、工具和会话持续协作的运行能力。

## 2. Agent Loop

Agent Loop 是持续调用模型、执行工具并反馈结果的运行循环，是 Harness 的核心组成部分。

以下是概念流程，不代表源码只有一个循环函数：

1. 组装任务、会话历史、指令及工具定义，向模型发起请求。
2. 接收模型输出的文本或工具调用。
3. 运行时根据权限、沙箱和审批策略处理工具调用；需要审批时等待决定。
4. 执行获准的工具，将结果加入上下文，再次请求模型。
5. 完成当前任务后返回结果，或因中断、错误等条件停止。

```text
用户任务 → 组装上下文 → 调用模型 → 输出最终结果
                         │
                         └→ 请求工具 → 权限与审批 → 执行工具
                                ↑                     │
                                └─ 再次调用模型 ← 反馈结果
```

模型决定尝试什么操作；运行框架负责推进循环，并在实际执行边界实施控制。模型输出工具调用不等于工具已获准执行。

## 3. 开源范围

官方 `openai/codex` 仓库采用 Apache 2.0 协议，核心实现主要使用 Rust。可查看的运行能力包括模型与工具交互、会话状态、流式事件、上下文管理，以及沙箱和审批处理。

模型权重、服务端推理实现和托管服务属于独立层次；运行框架开源不能推导出这些部分也开源，也不能推导出所有产品入口的界面均已开源。

## 4. 集成方式

| 入口 | 适用场景 |
|---|---|
| `codex exec` | 脚本、CI 与非交互任务 |
| Codex SDK | 在应用代码中以编程方式发起和管理任务 |
| Codex app-server | 需要会话、流式事件、中断和审批交互的产品集成 |

这些方式复用 Codex 的运行能力；应用仍需提供自己的业务上下文、工具和界面。

## 5. 源码与参考资料

- [Codex 官方仓库](https://github.com/openai/codex)：代码、构建说明与许可证。
- [核心运行代码](https://github.com/openai/codex/tree/main/codex-rs/core)：阅读执行循环及相关运行能力的入口；具体模块组织随版本变化。
- [Codex as a platform: build on the open agent harness](https://developers.openai.com/blog/codex-as-a-platform)：开源范围、Agent Loop 与集成方式的官方说明。
