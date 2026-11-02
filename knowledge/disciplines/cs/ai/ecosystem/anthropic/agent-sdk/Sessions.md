# Claude Agent SDK 会话管理

会话（Session）是 Claude Agent SDK 在执行期间积累的对话历史，包含用户提示词、工具调用、工具结果和 Agent 回复。SDK 默认自动将历史写入磁盘，供后续追问、进程重启后继续或分叉探索使用。本文依据官方文档整理，核对日期为 2026-11-02。

## 1. 会话与文件状态

会话持久化保存对话，不保存或恢复文件系统。恢复会话可以重新利用此前的分析和决策，但不代表磁盘文件仍与历史记录中的状态一致。

Fork 复制的是对话历史。若多个会话在同一个工作目录编辑文件，它们仍然修改同一份文件；需要文件快照与回退时，应使用文件检查点机制。

## 2. Continue、Resume 与 Fork

| 方式 | 会话选择与行为 | 适用场景 |
|---|---|---|
| Continue | 查找当前目录中最近的会话，并继续追加历史 | 单一对话的连续调用或重启后继续 |
| Resume | 根据指定会话 ID 继续追加历史 | 多用户应用、多个会话并存，或返回较早的会话 |
| Fork | 复制已有会话历史，创建具有独立 ID 的新会话；后续历史分别积累 | 保留原对话，尝试另一种方案 |

Python 的 `ClaudeSDKClient` 会在同一客户端实例的多次 `query()` 调用之间自动维持会话，不需要手动传递会话 ID。调用 `receive_response()` 消费当前查询的消息。

TypeScript 使用后续 `query()` 调用中的 `continue: true` 继续当前目录最近的会话。由于它通过目录选择最近会话，多用户或多会话应用应显式维护 ID，使用 Resume 避免选择错误。

## 3. 配置与会话 ID

| 操作 | Python | TypeScript |
|---|---|---|
| 继续最近会话 | `continue_conversation=True` | `continue: true` |
| 恢复指定会话 | `resume=session_id` | `resume: sessionId` |
| 从指定会话分叉 | `resume=session_id, fork_session=True` | `resume: sessionId, forkSession: true` |

以上字段传入 `query()` 的选项对象，即 Python 的 `ClaudeAgentOptions` 或 TypeScript 的 `options`。

会话 ID 可从结果消息的 `session_id` 字段读取，成功和错误结果均包含该字段。也可以更早从初始化消息获取：TypeScript 使用 init `SystemMessage` 的直接字段，Python 则从 `SystemMessage.data` 读取。

连接或进程故障可能导致没有结果消息。如果需要在此类故障后恢复，应尽早取得并持久保存会话 ID，不能只依赖正常结束时的结果消息。

### 3.1 Python 恢复示例

```python
from claude_agent_sdk import ClaudeAgentOptions, query


async def continue_analysis(session_id: str):
    async for message in query(
        prompt="根据之前的分析实施重构",
        options=ClaudeAgentOptions(
            resume=session_id,
            allowed_tools=["Read", "Edit", "Write", "Glob", "Grep"],
        ),
    ):
        print(message)
```

应用需要提供先前捕获的会话 ID，并通过异步入口调用此函数。单次 `query()` 在产生错误结果后可能继续抛出异常，应用应处理异常后再决定是否恢复；调整轮次或费用上限后，也可以继续原会话。

## 4. 存储与跨主机恢复

默认会话记录位于 `~/.claude/projects/<encoded-cwd>/*.jsonl`。设置 `CLAUDE_CONFIG_DIR` 后，位置改为该目录下的 `projects/`；工作目录名称会编码，较长名称还会截断并追加哈希。

按 ID 恢复仍要求会话记录存在于当前机器。跨机器、临时容器或无服务器环境可以采用以下方式：

- 配置 `sessionStore`（TypeScript）或 `session_store`（Python）适配器，将会话记录同步到外部存储。存储查找键与工作目录有关，恢复时需要匹配原始 `cwd`。
- 保存并迁移会话 JSONL 文件，再按 ID 恢复。官方说明，CLI v2.1.223 起可跨项目目录查找 ID；捆绑旧 CLI 的 SDK 仍受原项目目录及其 Git worktree 范围限制。
- 将分析结果、决策和文件差异保存为应用状态，作为新会话的输入，减少对完整会话文件迁移的依赖。

迁移对话记录不会自动迁移代码、依赖、运行进程或外部资源；应用还需准备相应执行环境。

## 5. 与故障恢复的边界

恢复会话可以继续已有对话，但官方会话文档未说明“文件已修改，工具结果尚未持久化就崩溃”这一窗口的事务保证，也未承诺自动回滚或避免重复副作用。

因此，不能把 Resume 理解为从任意机器指令位置继续执行。对结果未知的文件写入或远程操作，应用仍需根据工具能力核对实际状态，再判断是否重试或补偿。

## 6. 官方文档

- [Work with sessions](https://code.claude.com/docs/en/agent-sdk/sessions)
- [Persist sessions to external storage](https://code.claude.com/docs/en/agent-sdk/session-storage)
