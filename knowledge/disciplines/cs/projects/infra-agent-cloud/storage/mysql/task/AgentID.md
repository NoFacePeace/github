# 任务关联 Agent

`agent_id` 保存任务使用的 [Agent](../agent/Agent.md) 主键，类型为可空无符号大整数。

```text
Agent.id ← Task.agent_id
```

一个 Agent 可被多个任务使用；该字段允许为空，不能假设每个任务都绑定 Agent。
