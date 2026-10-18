# Runtime 函数

本目录整理 Go runtime 的内部函数，按函数记录职责、执行过程及调用场景。具体实现以所用 Go 版本为准，这些函数不是面向业务代码的公开 API。

## 1. 挂起与唤醒

- [gopark](gopark.md)：提交当前 goroutine 的等待状态，并进入调度。
- [goready](goready.md)：将等待中的 goroutine 置为可运行，等待后续调度。

## 2. 抢占

- [gopreempt_m](gopreempt_m.md)：在 `g0` 栈上处理普通抢占，将目标 G 重新排队并进入调度。
- [preemptPark](preemptPark.md)：在 `g0` 栈上响应停止请求，供运行时取得目标 G 的暂停控制权。

## 3. 暂停与恢复

- [suspendG](suspendG.md)：暂停目标 G 并取得栈等状态的检查权限。
- [resumeG](resumeG.md)：释放检查权限，并按暂停结果恢复运行资格。

## 4. 栈切换

- [systemstack](systemstack.md)：临时切到系统栈执行函数，完成后返回原执行上下文。
- [mcall](mcall.md)：保存当前 G 的恢复现场，切到 `g0` 并交由目标函数处理后续执行。

## 5. GC 后台任务

- [bgsweep](bgsweep.md)：后台清扫 goroutine 的入口，负责领取清扫工作、挂起及等待后续周期。
