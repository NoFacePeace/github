# gopreempt_m

`gopreempt_m` 是 Go runtime 处理普通抢占的内部函数，定义于 `src/runtime/proc.go`。它在当前 M 的 `g0` 栈上执行，让被抢占的 G 重新排队，并进入调度。

## 1. 调用入口

异步抢占满足安全条件时，信号处理路径调整返回现场，使目标 G 执行 `asyncPreempt`，保存寄存器后进入 `asyncPreempt2`。

`asyncPreempt2` 根据目标 G 的 `preemptStop` 标志选择后续路径：

```go
if gp.preemptStop {
    mcall(preemptPark)
} else {
    mcall(gopreempt_m)
}
```

`mcall` 保存当前 G 的恢复现场，切换到同一个 M 的 `g0` 栈，并将被抢占的 G 作为参数交给 `gopreempt_m`。

sysmon 因 G 长时间运行而发起的普通抢占，通常走这一分支。具体选择取决于目标 G 的标志，而不是信号种类。

## 2. 执行过程

`gopreempt_m` 通过 `goschedImpl` 完成让出执行权的核心处理，大致过程为：

1. 将目标 G 从 `_Grunning` 改为 `_Grunnable`。
2. 解除 G 与当前 M 的执行绑定。
3. 将目标 G 加入全局可运行队列。
4. 进入 `schedule`，选择后续执行的 G。

具体实现、跟踪记录和调度资源唤醒逻辑以所用 Go 版本为准。

## 3. 恢复执行

目标 G 仍具备运行资格，可以再次被调度，也可能换到另一个 M 上执行。恢复后，异步抢占路径还原寄存器等现场，回到原来被打断的位置继续运行。

让出执行权不保证一定换成其他 G。如果没有其他可运行工作，目标 G 也可能很快再次执行。

## 4. 与 preemptPark 的区别

| 函数 | 目的 | 目标 G 的状态 |
| --- | --- | --- |
| `gopreempt_m` | 普通抢占，让其他工作获得执行机会 | `_Grunnable`，重新排队 |
| `preemptPark` | 响应停止请求，供运行时取得暂停控制权 | `_Gpreempted`，需由运行时后续恢复运行资格 |

两者都在 `g0` 栈上执行并进入调度。`preemptPark` 常见于 `suspendG` 请求暂停正在运行的目标 G，例如为了安全扫描栈；不能将所有 GC 抢占都归为这一分支。
