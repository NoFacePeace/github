# 查询 Session

Redis 调度记录为 `queued` 时，调度器调用 Server 的任务处理逻辑，同步当前排队状态。此过程先查询业务记录。

## 1. 定位记录

根据 Task UUID 查询所属任务；若未找到适用的远程任务，本次业务同步直接返回。

找到任务后，在数据库事务中根据 Session UUID 和 Task 主键查询 [Session](../../../../../../storage/mysql/session/Session.md)，并对记录加悲观写锁。预期存在的 Session 未找到时会抛出错误，不能继续当作正常记录派发。

## 2. 判断状态

检查 [status](../../../../../../storage/mysql/session/Status.md)：

- `running`：继续按条件更新调度扩展信息。
- 非 `running`：不直接写入新的排队信息，后续按当前状态尝试取消排队或执行收尾。

`running` 表示本轮尚未结束，不代表 Agent 已启动。还需结合停止请求标记判断是否需要停止处理。

## 3. 后续步骤

Session 仍为 `running` 时，进入[更新扩展信息](UpdateSessionExtraInfo.md)，按条件保存调度状态副本。
