# sysmon

sysmon 是 Go 运行时的系统监控机制，核心循环位于 `src/runtime/proc.go` 的 `sysmon()`。它由专门的 M 执行，不需要持有 P，用于在普通调度工作之外检查运行时状态、请求抢占和唤醒后台工作。

## 1. 执行方式

运行时启动过程中，通过 `newm(sysmon, nil, -1)` 等路径启动系统监控线程。sysmon 在该 M 的系统栈上执行，不是通过普通 `go` 语句启动并放入运行队列的业务 G。

它与 netpoller 的执行方式不同：sysmon 有专门执行监控循环的 M，netpoller 则可由不同运行时路径上的 M 调用。sysmon 也会调用网络轮询，但不是唯一的调用者。

## 2. 主要职责

以下按 Go 1.27.2 的实现整理，具体条件可能随版本变化。

| 职责 | 行为 | 主要入口 |
| --- | --- | --- |
| 请求抢占 | 发现 P 的调度进度长时间未变化时，对正在运行的 G 请求抢占 | `retake()`、`preemptone()` |
| 回收调度资源 | 检查系统调用等状态，在满足条件时收回 P，使其他 M 能继续执行可运行 G | `retake()` |
| 网络轮询兜底 | 满足时间与轮询状态条件时，非阻塞检查网络事件并注入就绪 G | `netpoll(0)`、`injectglist()` |
| 检查时间驱动的 GC | 时间触发条件满足时，唤醒负责启动 GC 的后台 G | `gcTriggerTime`、`forcegc` |
| 唤醒内存归还工作 | 收到相应请求时唤醒 scavenger | `scavenger.wake()` |
| 检查 GOMAXPROCS 更新 | 开启自动更新时，周期性检查是否需要调整 P 数量 | `sysmonUpdateGOMAXPROCS()` |
| 输出调度跟踪 | 配置 `GODEBUG=schedtrace` 时，按条件输出调度状态 | `schedtrace()` |

sysmon 发出抢占请求不等于 G 立即停止；实际响应还取决于安全点、运行时状态和平台支持。它也不直接在监控循环中完成整轮 GC。

## 3. 检查与休眠节奏

### 3.1 自适应休眠

Go 1.27.2 的监控循环从约 20 微秒的休眠开始。连续没有唤醒工作时，逐步增加休眠时间，常规循环的上限为 10 毫秒。

如果所有 P 都空闲或运行时处于 STW 等状态，且未启用调度跟踪，sysmon 还可能进入更长的休眠，并结合计时器期限和唤醒通知恢复执行。

因此，sysmon 不是严格每隔固定时间执行一次的定时任务。

### 3.2 抢占检查

`retake()` 检查 P 的 `schedtick` 等进度信息。该版本的 `forcePreemptNS` 为 10 毫秒，当同一调度周期持续过久时，会请求抢占。

这不是对每个 G 的精确时间片保证。通过 `runnext` 连续执行的 G 可能共享调度周期，实际检查和抢占响应也存在延迟。

### 3.3 网络轮询检查

轮询器已初始化、`sched.lastpoll` 非零且距今超过 10 毫秒时，sysmon 调用 `netpoll(0)` 做非阻塞检查。`lastpoll` 为零通常表示已有阻塞轮询者。

该检查用于提供额外的进度保障，不是每个 M 都每隔 10 毫秒调用一次 `epoll_wait`。

### 3.4 GOMAXPROCS 更新检查

开启 `GODEBUG=updatemaxprocs` 对应的自动更新行为时，监控循环最多每秒检查一次更新。需要调整时，它会唤醒辅助 G 完成相关操作，因为 sysmon 不能直接执行 STW。

## 4. 源码入口

- `proc.go`：`runtime.main()` 中的启动路径、`sysmon()`、`retake()`、`sysmonUpdateGOMAXPROCS()`。
- `preempt.go`、`proc.go`：抢占请求与响应逻辑。
- `mgc.go`：GC 触发条件及后台 GC 启动逻辑。
- [Go 1.27.2 的 proc.go](https://github.com/golang/go/blob/go1.27.2/src/runtime/proc.go)：上述监控循环的版本固定源码。

## 5. 相关机制

- [Scheduler](scheduler/Scheduler.md)：goroutine 调度与抢占。
- [netpoller](netpoller.md)：I/O 事件获取、等待与唤醒。
- [GC](gc/GC.md)：垃圾回收的阶段与触发机制。

[返回 Runtime 目录](Runtime.md)
