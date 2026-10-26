# 执行所属任务

`task_id` 保存本轮执行所属的 [Task](../task/Task.md) 的主键。

```text
Task.id ← Session.task_id
```

一条 Session 属于一个 Task，一个 Task 可以对应多条 Session。
