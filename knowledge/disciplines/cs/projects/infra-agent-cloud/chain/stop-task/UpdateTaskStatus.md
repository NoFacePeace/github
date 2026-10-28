# 更新 Task 状态与返回请求

## 1. 处理内容

Task 停止入口逐个提交目标 Session 的停止请求后，统计目标轮次之外是否还有 running Session。若有，保留 Task 为 running；否则将 [Task](../../storage/mysql/task/Task.md) 更新为 stopped 并返回。

代码在此统计时排除本次目标 Session，因为启动阶段的停止可能在 Session 终态落库前返回。因此可以出现 Task 已 stopped、目标 Session 暂时仍 running 的窗口。

这是接口侧分支，与 Runner 异步停止和终态回传并行推进；Stop 接口成功不保证实际进程已退出，也不保证马上可以 Continue。

## 2. 后续流程

等待取消结果或执行终态落库后，以 Session 的实际状态判断本轮是否结束。
