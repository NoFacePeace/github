# runnerDispatch

`runnerDispatch` 是 Session 表 `extra_info` JSON 中的一个属性，不是独立数据库列。它保存本轮执行的调度状态副本，供任务查询、展示与执行控制使用。

## 1. 内容示例

以下标识与时间均为示例：

```json
{
  "taskUuid": "task-001",
  "sessionUuid": "session-001",
  "jobId": "job-001",
  "state": "queued",
  "queuedAt": 1800000000000,
  "deadline": 1800000600000
}
```

该对象位于 `extra_info.runnerDispatch`。后续可补充排队时长和错误原因，不保存加密执行输入本体。

## 2. 状态含义

| state | 含义 |
| --- | --- |
| queued | 等待调度。 |
| dispatching | 正在派发。 |
| dispatched | 已派发，不代表执行已完成。 |
| failed | 调度失败。 |
| cancelled | 调度已取消。 |

## 3. 同步时机

调度器将当前调度状态交给任务服务。任务服务查询 Task 与 Session，在事务中锁定 Session，保留其他扩展信息后更新本属性。

若 Session 已不在运行状态，不直接写入新的调度信息，而是按需取消排队或进行收尾。已有状态已超出排队阶段时，延迟到达的 `queued` 同步不能将其覆盖回排队状态。

## 4. 使用目的

任务查询可以据此区分正在排队与已进入派发等阶段。例如 Session 主状态仍为 `running`，本属性的 `state` 为 `queued`，表示本轮尚未结束、仍在等待 Runner。

它是调度状态在业务记录中的副本，实际调度处理仍依据 Redis 调度记录及相关协调状态进行。
