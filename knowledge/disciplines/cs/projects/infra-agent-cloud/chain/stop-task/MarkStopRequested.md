# 写入停止请求标记

## 1. 处理内容

在 MySQL 事务中按 Session 主键加写锁、重新读取，只有仍为 running 时写入 `extraInfo.runnerStopRequested = true`。已是终态时不重写该标记。

字段位置见 [Session 扩展信息](../../storage/mysql/session/ExtraInfo.md)。该标记用于后续派发、执行恢复和终态处理识别用户停止意图，本身不等于将 Session 状态改为 stopped。

## 2. 后续流程

[尝试取消排队](CancelPendingDispatch.md)。
