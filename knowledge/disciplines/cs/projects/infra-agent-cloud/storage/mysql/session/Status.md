# status

`status` 是 Session 表中本轮执行的整体状态，类型为 `varchar(20)`，默认值为 `running`。

## 1. 状态取值

| 值 | 含义 |
| --- | --- |
| running | 本轮仍在处理，可能处于准备、排队或实际执行阶段。 |
| completed | 本轮执行完成。 |
| failed | 本轮执行失败。 |
| stopped | 本轮已被标记为停止。 |

`running` 不保证 Agent 进程已经启动，`stopped` 也不能单独证明远端进程已经退出，仍需结合停止指令与执行事件判断。

## 2. 调度状态同步时的作用

同步排队状态时，服务端重新查询 Session。若 `status` 为 `running`，可继续写入调度扩展信息；若不是，则不再直接写入新的排队状态，并按需取消排队或处理执行收尾。

## 3. 与调度状态的区别

`queued`、`dispatching`、`dispatched` 等是调度记录的状态，不是本字段的取值。Session 可同时具有 `status = running` 和扩展调度信息中的 `state = queued`，表示本轮尚未结束、当前仍在等待 Runner。
