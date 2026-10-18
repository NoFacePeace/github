# Sysmon 休眠

sysmon 通过短时间的线程休眠和满足条件时的深度休眠控制监控频率。它不持续忙等，也不是严格按固定周期执行。本文按本地 Go 1.26.3 的 `src/runtime/proc.go` 整理，具体条件以所用版本为准。

## 1. 常规自适应休眠

`sysmon()` 在每轮循环开始时调用 `usleep(delay)`，参数单位为微秒：

| 条件 | 休眠时间调整 |
| --- | --- |
| `idle == 0` | 从 20 微秒开始 |
| `idle > 50` | 每轮将 delay 翻倍 |
| delay 超过 10,000 微秒 | 限制为 10 毫秒 |

`idle` 记录连续没有通过 `retake` 推进回收或抢占的循环次数。有相关进展时会重置，使后续检查重新变得频繁。

这套机制根据监控进展调整频率，而不是为每个 G 创建定时器。`usleep` 是线程级休眠，没有与 `sysmonnote` 配套的通知等待协议；常规休眠结束后，线程继续检查运行时状态。

## 2. 深度休眠条件

常规短休眠之后，代码判断是否适合进入更长的等待：

```go
debug.schedtrace <= 0 &&
    (sched.gcwaiting.Load() || sched.npidle.Load() == gomaxprocs)
```

含义是：未启用调度跟踪，并且运行时正在协调 STW，或所有 P 都空闲。随后加 `sched.lock`，再次检查相关状态，避免只依据加锁前的观察进入等待。

如果还有活跃的 P，sysmon 需要继续监控长时间执行、系统调用和网络事件等状态，不能仅因为业务 G 大多在等待就进入深度休眠。

## 3. 等待时长与执行过程

### 3.1 如何计算超时

sysmon 通过 `timeSleepUntil()` 获取各 P 上最近的计时器期限 `next`。如果期限已到，不进入这次深度休眠；否则计算：

```text
sleep = min(next - now, forcegcperiod / 2)
```

即使没有较近的业务计时器，也会限制最长等待时间，以便后续检查时间驱动的 GC 等工作。实际线程恢复时间还受操作系统调度影响。

### 3.2 如何等待

```text
加锁并确认休眠条件
    ↓
设置 sched.sysmonwait = true
    ↓
解锁
    ↓
notetsleep(&sched.sysmonnote, sleep)
    ├─ 其他运行时路径发送通知，提前返回
    └─ 等待超时，返回
    ↓
重新加锁，清除等待标记和 note
    ↓
继续监控循环
```

sysmon 使用的是带超时的 note 等待，不是 `time.Sleep`、`time.Timer` 或 `time.Ticker`。它参考业务 timer 的期限决定等待多久，不为自己的循环注册 Go timer。

sysmon 原本就不持有 P，因此它休眠时不存在先释放 P、醒来再获取 P 才能继续监控的问题。

## 4. 谁会唤醒 sysmon

其他 M 执行特定运行时路径时，会检查 `sched.sysmonwait`，需要时通过 `notewakeup(&sched.sysmonnote)` 通知它。

| 场景 | 主要入口 | 目的 |
| --- | --- | --- |
| STW 结束 | `startTheWorldWithSema` | 恢复监控 |
| G 进入普通系统调用 | `reentersyscall`、`entersyscallWakeSysmon` | 恢复检查，后续可能需要收回 P |
| 系统调用返回并通过相关路径获取空闲 P | `exitsyscallTryGetP`、`exitsyscallNoP` | P 重新活跃，恢复监控 |

以系统调用进入路径为例，代码有明确判断：

```go
if sched.sysmonwait.Load() {
    systemstack(entersyscallWakeSysmon)
}
```

辅助函数加锁后再次确认等待标记，清除标记并发送 note 通知。它判断的是 sysmon 是否正在深度休眠，不是预判系统调用是否耗时，也不代表要求立即收回某个 P。

G 的创建或普通唤醒不统一通知 sysmon。相关路径通常通过 `wakep` 等机制安排工作 M，这与唤醒监控线程是不同的动作。

## 5. 超时与通知的区别

`notetsleep` 的返回值区分收到通知和等待超时。该版本收到通知后会将 `idle` 重置为 0、`delay` 重置为 20 微秒，让监控重新变得频繁；仅因等待超时返回时，不做同样的重置。

业务计时器期限可以使等待超时，不要求另一个 M 在那个时刻显式发送通知。sysmon 醒来后继续监控，也不意味着业务 G 当场执行；业务 G 的恢复仍由计时器处理和调度流程完成。

## 6. 源码入口

- `src/runtime/proc.go`：`sysmon`、`entersyscallWakeSysmon`、系统调用返回及 STW 结束时的通知路径。
- `src/runtime/time.go`：`timeSleepUntil`，获取计时器期限。
- `src/runtime/lock_*.go`：note 等待与唤醒的平台实现。
