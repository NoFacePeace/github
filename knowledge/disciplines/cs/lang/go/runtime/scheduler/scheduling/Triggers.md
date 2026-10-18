# 调度触发

当前 G 等待、主动让出执行权、响应抢占或结束时，相关运行时路径会先处理它的状态，再进入调度。M 启动及部分系统调用返回路径也会进入调度。

## 1. 常见触发路径

| 场景 | 典型路径 | 原 G 的处理 |
| --- | --- | --- |
| [主动让出](Yield.md) | `runtime.Gosched` → `mcall(gosched_m)` → `goschedImpl` → `schedule` | 改为可运行，重新排队 |
| [等待条件](Waiting.md) | `gopark` → `mcall(park_m)` → `schedule` | 提交等待状态，等待对应机制唤醒 |
| [异步普通抢占](AsyncPreemption.md) | `asyncPreempt2` → `mcall(gopreempt_m)` → `goschedImpl` → `schedule` | 改为可运行，重新排队 |
| [抢占停止请求](PreemptionStop.md) | `asyncPreempt2` → `mcall(preemptPark)` → `schedule` | 进入 `_Gpreempted`，等待暂停请求方接管 |
| [G 执行结束](GoroutineExit.md) | `goexit` → `goexit1` → `mcall(goexit0)` → `schedule` | 清理执行资源，进入结束状态 |
| [M 启动](ThreadStart.md) | `mstart` → `mstart1` → `schedule` | 在具备所需调度资源后开始寻找工作 |
| [系统调用返回慢路径](SyscallReturn.md) | `exitsyscall` → `mcall(exitsyscall0)` → `schedule` | 无法立即恢复执行时，按协议重新安排原 G |

表中是典型分支，并非每次进入入口函数都必然调用 `schedule`。例如提交挂起失败可以直接恢复原 G，系统调用返回也可能直接取得 P 并继续执行。M 的具体启动入口和辅助调用随平台与版本变化。

## 2. 协作式抢占

请求设置、检查点和系统栈切换的详细过程参见[协作式抢占](CooperativePreemption.md)。

运行时将 `stackguard0` 设为 `stackPreempt` 后，G 执行到具有相应栈检查的函数入口，会触发 `morestack`。该入口保存现场并切到 `g0` 栈，由 `newstack` 识别和处理抢占请求。

```text
函数入口栈检查
    ↓
morestack → newstack（g0 栈）
    ├── 普通抢占 → gopreempt_m → goschedImpl → schedule
    └── 停止请求 → preemptPark → schedule
```

这条路径已通过 `morestack` 进入系统栈，无需再用 `mcall` 切一次。触发栈检查不代表一定扩栈；`newstack` 会区分抢占请求与实际栈空间需求。

## 3. 暂时不适合抢占

异步抢占信号到来时，如果当前位置不是安全点或运行时状态禁止抢占，信号处理路径不会注入抢占跳转，目标继续执行。后续仍可通过协作式检查或再次异步请求响应抢占。

协作式检查也需要确认能否安全抢占。处于不能安全中断的关键操作时，运行时会按协议暂缓，等待后续机会。因此，提出抢占或 STW 请求不等于立刻进入调度。

## 4. 唤醒与进入调度的区别

唤醒一个 G 只是让它获得运行资格，不保证调用方立即进入调度，也不保证目标 G 当场执行。创建 G、channel 唤醒和网络就绪等事件可以产生可运行工作，但不应都理解为当场调用 `schedule`。
