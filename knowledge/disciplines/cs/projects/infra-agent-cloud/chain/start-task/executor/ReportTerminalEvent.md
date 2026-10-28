# 上报终态事件

Executor 根据停止标记和退出码，发出本轮执行的终态事件。

## 1. 事件选择

| 条件 | 事件 |
| --- | --- |
| 已停止或中止信号已触发 | `stopped` |
| 未停止且退出码为 0 | `completed` |
| 未停止且退出码不为 0 | `failed` |

正常退出路径携带解析后的结果、用量等适用信息。执行过程抛出异常时，异常处理分支会按停止状态发出 `stopped`，或携带错误信息发出 `failed`。

## 2. 与 Runtime、Server 的关系

Executor 将事件交给外层 Runtime，由 Runtime 负责传输。Server 消费后更新业务状态；Executor 不直接更新 Task、Session 表，也不直接写事件 Redis key。

停止路径返回码为 `130`，正常路径返回引擎退出码，异常失败路径返回 `1`。

## 3. 下一步

结束路径进入[关闭执行资源](CloseExecutionResources.md)。
