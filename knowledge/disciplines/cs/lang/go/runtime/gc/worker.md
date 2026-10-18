# GC 标记 worker

后台标记 worker 是 Go runtime 专门执行 GC 标记工作的 goroutine，运行函数为 `gcBgMarkWorker`。它需要获得 P 才能执行，不独占一条操作系统线程。具体实现以所用 Go 版本为准。

## 1. 创建时机

`gcStart` 在启动 GC 的准备过程中，通过 `gcBgMarkStartWorkers` 等逻辑补齐需要的 worker，通常位于清扫终止的 STW 之前。

运行时通常按 P 的数量准备 worker。已有 worker 可以跨 GC 周期复用，不是每轮重新创建；首次准备或 P 数量增加等情况下，可能需要补充创建。

## 2. 挂起与可用池

worker 没有安排标记工作时，在 `gcBgMarkWorker` 中通过 `gopark` 进入 `_Gwaiting`，保留自己的栈和执行现场，让原 M 可以执行其他 G。

运行时在提交挂起的回调中，将关联 worker 的记录放入 [gcBgMarkWorkerPool](../variables/gcBgMarkWorkerPool.md)。池中保存的是节点记录，节点通过 `gp` 指向 worker G。

休眠的 worker 不在 P 的普通可运行队列中，也不会占用一条线程持续等待。

## 3. 调度与执行

调度器结合 GC 进度、待处理工作及执行资源，从可用池中选择 worker。选中后将对应 G 从 `_Gwaiting` 转为 `_Grunnable`，交给调度流程执行，不一定先放入 P 的普通运行队列。

worker 执行根扫描、对象扫描等标记工作。结束当前执行片段后再次挂起，关联记录回到可用池，供后续调度。

```text
创建 worker G
    ↓ 提交挂起并发布可用节点
worker 池
    ↓ 调度器选择执行
执行标记工作
    ↓ 当前执行片段结束，再次挂起
回到 worker 池
```

只有一个 P 时，worker 与业务 G 通过调度交错执行；多个 P 时可以并行执行，但并不保证所有 worker 始终同时运行。

## 4. 与标记辅助的区别

| 对比 | 后台标记 worker | 标记辅助 |
| --- | --- | --- |
| 执行者 | 专门的 GC goroutine | 正在分配内存的业务 G |
| 调度方式 | 由运行时按标记策略安排 | 在业务分配路径中参与 |
| 是否使用 worker 池 | 是 | 不从池中领取 worker 代替自己执行 |

两者都推动标记进度，也都可能消耗业务可用的 CPU 时间。标记辅助还可能直接增加分配操作的延迟。

## 5. 源码入口

- `src/runtime/mgc.go`：worker 创建、执行循环、挂起及可用池管理。
- `src/runtime/mgcpacer.go`：worker 选择与标记调度策略。
- `src/runtime/mgcmark.go`：标记工作处理及辅助标记。
