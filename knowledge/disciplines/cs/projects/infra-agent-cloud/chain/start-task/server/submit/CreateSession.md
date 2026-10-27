# 创建 Session

Task 保存后，创建并持久化首个 [Session](../../../../storage/mysql/session/Session.md)，记录本轮输入及执行状态。

## 1. 生成执行轮次标识

生成 Session UUID，用于关联本轮执行请求、进度事件和结果。它与数据库自增主键及底层 Agent 引擎的会话标识分别使用。

## 2. 组织执行记录

| 字段 | 内容 |
| --- | --- |
| uuid | 新生成的 Session UUID。 |
| task_id | 已保存 Task 的数据库主键。 |
| prompt | 准备后的首次消息，作为本轮执行输入。 |
| status | 本轮执行的初始状态。 |
| creator | 本轮执行的发起者。 |

初始状态即使标记为运行中，也不代表执行进程已经启动。调度与实际启动仍在后续步骤完成。

## 3. 保存与关联

将 Session 写入数据库，通过 `task_id` 关联所属 Task：

```text
Task.id ← Session.task_id
```

发起新任务时，Task 与首个 Session 分别形成一条记录；之后启动新的执行轮次时，可复用 Task 并新增 Session。

## 4. 后续步骤

任务与首轮执行记录保存后，进入[准备内部工具上下文](PrepareInternalToolContext.md)，按需注入工具配置，再构建远程执行输入，为后续提交调度做准备。后续执行事件携带任务和执行轮次标识，用于更新对应记录的进度、状态与结果。

接口可先返回 Task UUID、Session UUID 和当前状态，供调用方查询或继续交互；这不表示任务已经完成。
