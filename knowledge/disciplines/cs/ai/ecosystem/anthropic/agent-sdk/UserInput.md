# 用户输入与审批

Claude Agent SDK 支持在任务执行期间向用户请求工具审批或提出澄清问题。应用接收请求、展示交互界面并返回用户决策，Agent 再继续执行。本文依据官方文档整理，核对日期为 2026-11-01。

## 1. 交互场景

| 场景 | 示例 | 触发方式 |
|---|---|---|
| 工具审批 | 是否允许执行命令、修改或删除文件 | 工具调用经过权限评估后需要批准 |
| 需求澄清 | 项目使用 PostgreSQL 还是 MySQL | Claude 调用 `AskUserQuestion` 工具 |

两类请求都通过 Python 的 `can_use_tool` 或 TypeScript 的 `canUseTool` 回调处理。回调等待期间，执行流程暂停，直到应用返回响应。这与 Agent 完成一轮普通聊天后等待下一条用户消息不同。

```text
Agent 执行任务
      ↓
工具需要批准 / Claude 调用 AskUserQuestion
      ↓
SDK 调用应用提供的权限回调
      ↓
应用展示问题，等待用户决策
      ↓
回调返回批准、拒绝或用户答案
      ↓
Agent 继续执行
```

SDK 提供请求与响应机制；终端提示、网页弹窗或移动端表单等界面由应用实现。

## 2. 工具审批

### 2.1 回调输入与响应

权限回调接收工具名称、工具输入，以及附加上下文。附加上下文可包含建议的权限更新；TypeScript 还提供用于取消等待的 `AbortSignal`，Python 对应的信号字段当前为预留字段。

| 决策 | Python 返回值 | TypeScript 返回值 |
|---|---|---|
| 批准 | `PermissionResultAllow(updated_input=...)` | `{ behavior: "allow", updatedInput: ... }` |
| 拒绝 | `PermissionResultDeny(message=...)` | `{ behavior: "deny", message: ... }` |

批准时可以保留或修改工具参数；拒绝时返回原因，Claude 会看到该消息并可能调整执行方式。应用也可以在用户选择后返回建议的权限更新，实现“本次允许”或“允许并记住规则”等交互。

### 2.2 与权限规则及 Hooks 的关系

`canUseTool` 不是每次工具调用都会触发的拦截器。前面的权限规则或权限模式已经批准的调用，通常不会进入该回调；某些模式也会直接拒绝调用而不询问用户。实际行为应以权限评估顺序为准。

如果需要对每次工具调用执行检查，应使用 `PreToolUse` Hook。Hook 可在工具执行前允许、拒绝或修改请求；权限回调则主要负责进入人工决策流程的请求。

## 3. 澄清问题

### 3.1 AskUserQuestion

Claude 需要进一步了解需求时，可以调用 `AskUserQuestion`。应用在权限回调中检查工具名称，并将其转交给专门的问题处理逻辑。

问题与选项由 Claude 生成，应用负责展示和收集答案。应用自身需要主动发起的其他问题，应由自己的业务逻辑另行处理。

如果通过 `tools` 列表限制可用工具，需要在其中包含 `AskUserQuestion`，否则 Claude 无法使用该工具提问。

### 3.2 问题与答案结构

每次调用支持 1–4 个问题，每个问题包含 2–4 个选项。

| 字段 | 含义 |
|---|---|
| `question` | 完整问题文本 |
| `header` | 简短标签，最多 12 个字符 |
| `options` | 选项数组，每项包含 `label` 和 `description` |
| `multiSelect` | 是否支持多选 |

应用通过批准响应中的更新输入返回原始 `questions` 和 `answers`。`answers` 的键是完整问题文本，值为选中的选项标签；多选可以返回标签数组或逗号加空格连接的字符串。

```json
{
  "questions": [
    {
      "question": "项目使用哪个数据库？",
      "header": "数据库",
      "options": [
        { "label": "PostgreSQL", "description": "使用 PostgreSQL" },
        { "label": "MySQL", "description": "使用 MySQL" }
      ],
      "multiSelect": false
    }
  ],
  "answers": {
    "项目使用哪个数据库？": "PostgreSQL"
  }
}
```

界面也可以提供自由文本输入，将用户填写的实际内容作为对应问题的答案，而非返回“其他”标签。如果用户关闭问题卡片并输入不对应某个问题的一般回复，可以使用可选的 `response` 字段；设置后 Claude 接收的是该一般回复，而非逐题答案列表。

## 4. 等待、恢复与其他输入方式

权限回调可以持续等待，但等待期间宿主进程需要保持运行。如果用户可能很久以后才回复，官方建议通过 `PreToolUse` Hook 返回 `defer` 决策，让进程可以退出，之后从已持久化的会话继续，而不是一直等待回调。

| 方式 | 适用场景 |
|---|---|
| 权限回调与 `AskUserQuestion` | 工具审批、Claude 发起的需求澄清 |
| 流式输入 | 用户执行中补充上下文、发送后续消息或改变任务方向 |
| 自定义工具 | 复杂表单、多步骤向导、外部审批系统或领域专用交互 |

当前通过 `Agent` 工具启动的子 Agent 不支持 `AskUserQuestion`。实际接入时还应检查 SDK 版本及语言差异；官方 Python 示例当前包含流式输入与维持流连接的 Hook 处理，应结合所用版本核对。

## 5. 参考资料

- [Handle approvals and user input](https://code.claude.com/docs/en/agent-sdk/user-input)
- [Configure permissions](https://code.claude.com/docs/en/agent-sdk/permissions)
- [Control execution with hooks](https://code.claude.com/docs/en/agent-sdk/hooks)
- [Defer a tool call for later](https://code.claude.com/docs/en/hooks#defer-a-tool-call-for-later)
- [Streaming input](https://code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode)
