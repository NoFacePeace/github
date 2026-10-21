# Claude 动态工作流

Dynamic Workflows 是 Claude Managed Agents 的多 Agent 编排功能。主 Agent 根据任务生成工作流程序，由服务端后台执行，程序调用多个 Agent、传递并汇总结果。它是一项 Agent 编排能力，不是独立的 Agent 产品。

信息核对日期：2026 年 10 月 12 日。官方于 2026 年 10 月 9 日发布该功能，当前处于 Beta；接口与限制以官方文档为准。

## 1. 执行机制

```text
用户通过 Managed Agents API 提交任务
                  ↓
主 Agent 判断是否需要工作流并生成编排程序
                  ↓
服务端启动 Workflow Run，后台执行程序
                  ↓
程序调用多个 Agent，收集结果并控制后续步骤
                  ↓
主 Agent 读取运行结果，回复用户
```

Workflow 是程序，Workflow Run 是一次程序执行。程序可以并行调用 Agent、传递结果、进行分支和循环，也可以处理失败或让失败终止运行。主 Agent 不必逐步参与每次调度，可以继续工作或结束当前轮次。

每个工作流 Agent 在独立的会话线程中运行，拥有自己的对话历史；同一会话中的线程共享沙箱文件。对话上下文独立不等于执行环境独立。

## 2. API 与启用方式

开发者通过 Managed Agents API 配置 Agent、创建 Session、发送用户消息并读取事件流。Agent 定义中的启用配置为：

```json
{
  "multiagent": {
    "type": "multiagent_20261001",
    "workflows": {
      "type": "enabled"
    }
  }
}
```

该类型默认启用 workflows 和 subagents。Managed Agents 请求需要 `managed-agents-2026-04-01` Beta Header，官方 SDK 自动设置该 Header。

启用功能后，由主 Agent 决定何时启动 Run。可以通过系统提示词或用户消息指导它，例如要求多份文档并行审查、单份文档直接处理。

当前文档没有提供客户端直接提交编排程序并创建 Run 的独立公开 API；只有会话主 Agent 能启动 Run。对外 API 与 Agent 内置的工作流启动能力属于不同层次。

## 3. 与 Subagent 的区别

| 方面 | Subagent | Dynamic Workflows |
|---|---|---|
| 流程控制 | 主 Agent 委派任务并决定下一步 | 生成的程序控制调用、阶段、分支和循环 |
| 结果传递 | 子 Agent 向主 Agent 报告 | Agent 向程序返回结果，再由程序传给后续任务 |
| 后续沟通 | 主 Agent 可向已有 Subagent 发送后续消息 | 主 Agent 不能向 Run 内线程发送后续消息 |
| 线程生命周期 | 可保留供后续使用，直到归档 | 服务端最迟在 Run 结束时归档 |

工作流适合包含大量可拆分任务、并行处理和多阶段复核的工作。每个 Agent 的模型调用都消耗 Token，应配置会话预算。

## 4. 状态与事件

调用方通过会话事件流中的 `workflow_run.*` 事件跟踪执行：

| 事件 | 含义 |
|---|---|
| `workflow_run.created` | Run 已创建，包含名称和声明的阶段 |
| `workflow_run.status_running` | 开始或恢复运行 |
| `workflow_run.status_idle` | 暂停，例如达到预算 |
| `workflow_run.phase_started` / `phase_ended` | 阶段开始或结束 |
| `workflow_run.status_ended` | Run 结束，返回完成、停止或错误状态 |
| `workflow_run.error` | 创建或运行错误，不能仅凭该事件认定 Run 已结束 |

`completed` 表示程序运行结束，不保证所有子任务成功或产物满足目标。需要结合线程事件和最终结果验证任务。

## 5. 已公开与未公开的实现细节

官方公开了 Agent 生成程序、后台执行、线程创建、结果传递与状态事件等运行机制，但没有在上述文档中明确说明程序语言、完整执行引擎或内部调度框架。

默认事件流不展示工作流代码。官方建议等 Run 结束且会话空闲后，要求主 Agent 原样输出启动时的工作流。不能据此假定底层使用某一种语言、Temporal 或其他特定框架。

通用编排模式参见 [Agent 编排](../../../agent/control/orchestration/Orchestration.md)。

## 6. 参考资料

- [官方发布记录](https://platform.claude.com/docs/en/release-notes/overview)
- [多 Agent 编排与启用配置](https://platform.claude.com/docs/en/managed-agents/multiagent-orchestration)
- [Workflow Run 运行机制、事件与限制](https://platform.claude.com/docs/en/managed-agents/workflow-runs)
