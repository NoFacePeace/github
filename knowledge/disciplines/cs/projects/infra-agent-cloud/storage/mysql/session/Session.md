# Session

Session 表保存任务中一轮执行的信息，例如本轮输入、执行状态、日志与结果的相关记录。

## 1. 生命周期

发起任务时创建首条 Session；继续任务并启动新一轮执行时，新增 Session。运行中补充消息也可能交给当前执行处理，不一定创建新记录。

## 2. 所属任务

Session 通过 [task_id](TaskID.md) 关联 Task 表的 `id`，每条 Session 属于一个 Task。

```text
Task.id ← Session.task_id
```

这里的 Session 表示平台的一轮执行，与底层 Agent 引擎自身的会话标识应分别理解。
