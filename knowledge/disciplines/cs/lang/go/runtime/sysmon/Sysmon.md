# Sysmon

sysmon 是 Go 运行时的系统监控机制，核心循环位于 `src/runtime/proc.go` 的 `sysmon()`。它由专门的 M 执行，不需要持有 P，用于在普通调度工作之外检查运行时状态、请求抢占和唤醒后台工作。

## 1. 执行方式

运行时启动过程中，通过 `newm(sysmon, nil, -1)` 等路径启动系统监控线程。sysmon 在该 M 的系统栈上执行，不是通过普通 `go` 语句启动并放入运行队列的业务 G。

创建 M、进入监控循环的调用链，以及不需要 P 的原因，参见 [Bootstrap](Bootstrap.md)。

它与 netpoller 的执行方式不同：sysmon 有专门执行监控循环的 M，netpoller 则可由不同运行时路径上的 M 调用。sysmon 也会调用网络轮询，但不是唯一的调用者。

## 2. 主要职责

监控内容、各项职责及后续处理方式参见[Sysmon 监控](Monitoring.md)。

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

休眠条件、超时计算及通知唤醒过程参见[Sysmon 休眠](Sleep.md)。

### 3.1 自适应休眠

监控循环从约 20 微秒的休眠开始。连续没有通过 `retake()` 推进回收或抢占时，逐步增加休眠时间，常规循环的上限为 10 毫秒。

如果所有 P 都空闲或运行时处于 STW 等状态，且未启用调度跟踪，sysmon 还可能进入更长的休眠，并结合计时器期限和唤醒通知恢复执行。

因此，sysmon 不是严格每隔固定时间执行一次的定时任务。

### 3.2 休眠方式与定时器

sysmon 不为自己的监控循环创建 Go `timer`，也不使用 `time.Timer` 或 `time.Ticker`。它依靠线程休眠和带超时的通知等待控制检查节奏：

| 方式 | 等待条件 | 用途 |
| --- | --- | --- |
| `usleep(delay)` | 等待一段时间，没有配套的 note 通知机制；参数单位为微秒 | 常规循环中的短休眠 |
| `notetsleep(&sched.sysmonnote, sleep)` | 等待通知或超时；超时参数单位为纳秒 | 满足条件时的较长休眠，可由其他 M 提前唤醒 |

较长休眠前，sysmon 通过 `timeSleepUntil()` 获取各 P 上最近的 timer 到期时间，将等待时长限制为 `min(最近到期时间 - now, forcegcperiod / 2)`；到期时间已过则不进入这次长休眠。它参考业务 timer 的期限来安排检查，不为此注册自己的 Go timer。等待超时由底层 OS 机制支持。

### 3.3 谁通知 sysmon

以下通知路径已对照本地 Go 源码 `714d9afb66` 核对。其他 M 执行相应运行时代码时，检查 `sched.sysmonwait`；确认 sysmon 正在长休眠后，通过 `notewakeup(&sched.sysmonnote)` 通知它。

| 时机 | 入口 | 目的 |
| --- | --- | --- |
| STW 结束，恢复运行 | `startTheWorldWithSema` | 恢复监控 |
| G 进入系统调用 | `entersyscallWakeSysmon` | 后续可能需要检查并回收系统调用中的 P |
| 系统调用返回，通过获取空闲 P 的路径恢复执行 | `exitsyscallTryGetP`、`exitsyscallNoP` | P 重新活跃，恢复监控 |

G 变为可运行本身不会统一通知 sysmon。创建或唤醒 G 的路径通常通过 `wakep()` 尝试唤醒或创建工作 M 来执行任务；这是唤醒工作线程，与通知 sysmon 是两件事。特定系统调用路径可能同时涉及 G 恢复执行和通知 sysmon，但不能据此理解为每次 G 就绪都会通知它。

没有收到通知时，`notetsleep` 也会在超时后返回；timer 到期不要求有其他 M 显式通知 sysmon。其返回值区分收到通知与等待超时，监控循环据此调整休眠状态。

### 3.4 抢占检查

系统调用中 P 的回收条件、状态协调及交接过程参见Sysmon P 回收。

`retake()` 检查 P 的 `schedtick` 等进度信息。该版本的 `forcePreemptNS` 为 10 毫秒，当同一调度周期持续过久时，会请求抢占。

这不是对每个 G 的精确时间片保证。通过 `runnext` 连续执行的 G 可能共享调度周期，实际检查和抢占响应也存在延迟。

请求发出后，G 如何通过函数入口检查或定向线程信号响应，以及内核、中断与单核调度如何参与，参见 Go Goroutine 抢占机制。

### 3.5 网络轮询检查

兜底检查的原因、调用条件及就绪 G 的调度过程参见Sysmon 网络轮询。

轮询器已初始化、`sched.lastpoll` 非零且距今超过 10 毫秒时，sysmon 调用 `netpoll(0)` 做非阻塞检查。`lastpoll` 为零通常表示已有阻塞轮询者。

该检查用于提供额外的进度保障，不是每个 M 都每隔 10 毫秒调用一次 `epoll_wait`。

### 3.6 GOMAXPROCS 更新检查

开启 `GODEBUG=updatemaxprocs` 对应的自动更新行为时，监控循环最多每秒检查一次更新。需要调整时，它会唤醒辅助 G 完成相关操作，因为 sysmon 不能直接执行 STW。

## 4. 源码入口

- `proc.go`：`runtime.main()` 中的启动路径、`sysmon()`、`retake()`、`sysmonUpdateGOMAXPROCS()`。
- `preempt.go`、`proc.go`：抢占请求与响应逻辑。
- `mgc.go`：GC 触发条件及后台 GC 启动逻辑。
- [Go 1.27.2 的 proc.go](https://github.com/golang/go/blob/go1.27.2/src/runtime/proc.go)：上述监控循环的版本固定源码。

## 5. 相关机制

- [Bootstrap](Bootstrap.md)：sysmon 的创建时机、专用 M 启动与无 P 执行机制。
- Go Goroutine 抢占机制：协作式与异步抢占、线程信号及单核执行过程。
- 时间条件触发：sysmon 检查时间条件并唤醒后台强制 GC 的过程。
