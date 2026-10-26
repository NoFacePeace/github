# 版本所属 Agent

`agent_id` 保存所属 [Agent](../agent/Agent.md) 的主键，类型为非空无符号大整数。

```text
Agent.id ← AgentVersion.agent_id
```

一个 Agent 可以对应多条版本记录，每条版本记录属于一个 Agent。字段关联不等同于数据库一定存在外键约束。
