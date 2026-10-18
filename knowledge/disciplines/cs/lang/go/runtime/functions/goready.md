# goready

`goready` 是 Go runtime 将等待中的 goroutine 置为可运行的内部函数，定义于 `src/runtime/proc.go`。唤醒只表示获得运行资格，不表示目标 G 立即执行。

## 1. 调用前提

调用方应先完成共享状态更新及等待结果记录，再让目标 G 变为可运行。`goready` 不负责传递 channel 数据或选择通信分支，也不负责查找等待队列中的目标节点。

例如，channel 直接配对通信时，调用方先处理数据、记录成功并释放 channel 锁，再通过等待节点的 `g` 字段找到需要唤醒的 G。

## 2. 唤醒过程

1. `goready` 通过 `systemstack` 在系统栈上调用内部的 `ready`。
2. `ready` 检查目标 G 的状态，将其从 `_Gwaiting` 改为 `_Grunnable`。
3. 将 G 加入可运行队列，并按需要唤醒调度资源来执行可运行工作。
4. 调度器选中目标 G 后，它才恢复运行。

具体入队位置与运行时策略有关，不能依赖唤醒顺序推断最终执行或完成顺序。

## 3. 与挂起的配合

```text
_Grunning
    ↓ gopark / park_m
_Gwaiting
    ↓ goready / ready
_Grunnable
    ↓ 调度器选中
_Grunning
```

目标 G 通常从先前 `gopark` 调用返回的位置继续，由对应等待机制完成结果检查和资源清理。

## 4. 相关函数

- [gopark](gopark.md)：提交当前 G 的等待状态并进入调度。
