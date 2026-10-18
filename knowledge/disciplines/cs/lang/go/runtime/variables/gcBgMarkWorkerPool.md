# gcBgMarkWorkerPool

`gcBgMarkWorkerPool` 是 Go runtime 管理可用后台标记 worker 的全局变量，定义于 `src/runtime/mgc.go`。它不是类型名，其类型为内部的 `lfstack`，用于组织可用 worker 的记录。具体字段及调度流程以所用 Go 版本为准。

## 1. 池中的节点

池中存放的是 `gcBgMarkWorkerNode` 记录，不是直接把 G 对象作为链表节点。其核心结构可简化为：

```go
// 简化示意，并非完整结构定义。
type gcBgMarkWorkerNode struct {
    node lfnode // 连接 lfstack 的节点
    gp   *g     // 关联的后台标记 worker G
}
```

`lfstack` 是运行时内部的无锁栈，提供节点入栈和出栈操作；节点还可以包含其他调度相关信息。

## 2. worker 的创建与挂起

`gcStart` 在启动 GC 的准备过程中，通过 `gcBgMarkStartWorkers` 等逻辑补齐所需 worker。通常在清扫终止的 STW 之前准备，已有 worker 可以在后续 GC 中复用。

worker 执行 `gcBgMarkWorker`，没有安排标记工作时，通过 `gopark` 进入 `_Gwaiting`。运行时在提交挂起的回调中发布可用节点，协调“已挂起”和“可以从池中领取”的顺序，避免节点被过早取走。

挂起的 worker 保留 G 对象、栈和恢复现场，不占用一条线程持续等待，也不在 P 的普通可运行队列中。

## 3. 领取与执行

调度器需要执行后台标记工作时，通过 `gcController.findRunnableGCWorker` 等路径从池中取得节点，找到对应 G，并根据标记工作及调度策略决定是否安排执行。

选中的 worker 从 `_Gwaiting` 转为 `_Grunnable` 后交给调度流程，不一定先进入 P 的普通运行队列。如果取出节点后发现暂时不适合执行，可以将节点放回池中。

worker 获得执行机会后消耗标记任务；结束当前执行片段后再次挂起，并将可用记录放回池中，以供后续调度。

```text
worker 提交挂起
    ↓ 发布可用节点
gcBgMarkWorkerPool
    ↓ 调度器领取并选择执行
关联的 G 变为可运行，执行标记
    ↓ 当前执行片段结束
再次挂起并回池
```

## 4. 与其他队列的区别

| 结构 | 保存的内容 | 用途 |
| --- | --- | --- |
| `gcBgMarkWorkerPool` | 可用 worker 的记录 | 找到可以安排标记工作的 G |
| P 的可运行队列 | 可运行的 G | 普通调度 |
| 标记工作缓冲 | 待扫描对象等标记工作 | 提供 worker 和标记辅助需要处理的任务 |

池中有可用 worker，不代表一定有待扫描对象，也不表示必须立即调度。标记辅助则由业务 G 自己执行，不是从该池中领取 worker 来替它执行。

## 5. 源码入口

- `src/runtime/mgc.go`：worker 池、节点、创建和运行循环。
- `src/runtime/mgcpacer.go`：后台标记 worker 的选择与调度策略。
- `src/runtime/lfstack.go`：内部无锁栈实现。
