# park_m

`park_m` 是 Go runtime 提交 goroutine 等待状态并转入调度的内部函数，定义于 `src/runtime/proc.go`。它由 `gopark` 通过 `mcall` 调用，在当前 M 的 `g0` 栈上执行，参数是准备挂起的原 G。

## 1. 从 gopark 到 park_m

`gopark` 在当前 G 的执行上下文中记录等待原因和提交回调，再调用 `mcall(park_m)`。`mcall` 保存原 G 的恢复现场，切到同一个 M 的 `g0` 栈，将原 G 的指针传给 `park_m`。

```text
原 G 栈：gopark
             ↓ mcall(park_m)
g0 栈：  park_m(原 G)
```

切换执行栈不创建新线程。`park_m` 执行时操作的是原 G 的调度状态，不是将 `g0` 自己挂起。

## 2. 提交等待状态

`park_m` 将原 G 从 `_Grunning` 改为 `_Gwaiting`，并通过 `dropg` 解除原 G 与 M 的执行绑定，再执行 `gopark` 登记的提交回调（如果有）。

等待队列由调用方管理，`park_m` 不负责创建 channel 等待节点。例如，channel 已在调用 `gopark` 前把 `sudog` 加入发送或接收等待队列。

## 3. 提交回调

回调在 G 已进入等待状态后执行，用于协调等待提交和资源释放。channel 的 `chanparkcommit` 在这里释放 channel 锁，使通信另一方取得锁后能够安全完成数据交接和唤醒。

如果回调返回失败，运行时撤销此次等待，将原 G 恢复为可运行并通过执行路径直接继续它，不进入正常的等待调度分支。

## 4. 进入调度

等待提交成功后，`park_m` 进入 [schedule](schedule.md)，让当前 M 寻找后续可运行工作。原 G 处于等待状态，不会仅因进入调度就自动重新排队。

```text
原 G：_Grunning → _Gwaiting
    ↓ 提交回调确认等待
schedule：选择其他可运行 G
```

没有可运行工作时，运行时可能让 M 等待，而不是让它持续忙等原 G 的条件。

## 5. 原 G 如何恢复

条件满足后，其他运行时路径通过 `goready` 等机制将原 G 置为可运行。调度器再次选中它时，恢复保存的执行现场，使其从此前 `gopark` 调用处继续执行。

原 G 不会从 `park_m` 的函数尾部直接返回；它由调度器通过保存的现场恢复，也可能换到另一个 M 上运行。具体辅助逻辑与状态检查以所用 Go 版本为准。
