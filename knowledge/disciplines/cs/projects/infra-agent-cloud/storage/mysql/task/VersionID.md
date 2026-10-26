# 任务关联版本

`version_id` 保存任务绑定的 [AgentVersion](../agent-version/AgentVersion.md) 记录主键，类型为可空无符号大整数。

```text
AgentVersion.id ← Task.version_id
```

该字段关联版本记录的 `id`，不是 `version` 版本号。多个任务可以使用同一版本记录。
