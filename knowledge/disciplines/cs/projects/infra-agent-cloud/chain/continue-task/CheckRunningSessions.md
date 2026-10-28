# 检查运行中的执行轮次

## 1. 处理内容

查询该 Task 下是否仍存在状态为 `running` 的 Session；存在时拒绝本次 Continue。状态含义见 [Session 状态](../../storage/mysql/session/Status.md)。

正常完成、失败、停止后都可继续，但实际以 Session 状态为准。点击 Stop 或 Task 已显示 stopped，不保证 Session 已退出 running；停止请求和终态处理存在时间差。

## 2. 下一步

进入[查询上一轮 Session](QueryPreviousSession.md)。
