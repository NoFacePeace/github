# Runtime

Go 的运行时负责 goroutine 调度、垃圾回收、内存分配、栈管理及网络轮询等基础机制。核心实现位于 Go 源码的 `src/runtime` 目录，调度器和 GC 都属于 `runtime` 包。

## 1. 源码组织

运行时按职责拆分为多个源码文件，调度器和 GC 没有分别拆成 `runtime/scheduler`、`runtime/gc` 等独立包。它们需要共享底层状态：GC 会暂停和恢复程序执行、扫描 goroutine 栈，内存分配也会触发 GC 和辅助标记。

| 子系统 | 主要源码文件 | 职责 |
| --- | --- | --- |
| 调度器 | `proc.go`、`preempt.go` | goroutine 调度、运行队列、工作窃取与抢占 |
| 核心结构 | `runtime2.go` | 定义 G、M、P 等运行时结构 |
| 垃圾回收 | `mgc.go`、`mgcmark.go`、`mgcsweep.go`、`mgcpacer.go` | GC 阶段管理、标记、清扫与节奏控制 |
| 内存分配 | `malloc.go`、`mheap.go`、`mcache.go`、`mcentral.go` | 对象分配、堆页管理与分配缓存 |
| 写屏障 | `mbarrier.go` | 维护并发标记期间的正确性 |
| 栈管理 | `stack.go` | goroutine 栈的分配、扩容与收缩 |
| 网络轮询 | `netpoll.go`、`netpoll_*.go` | 与操作系统 I/O 事件机制协作，唤醒等待中的 goroutine |
| 内存归还 | `mgcscavenge.go` | 将空闲物理内存归还操作系统 |

具体文件组织和函数实现可能随 Go 版本变化，阅读时应以使用版本的源码为准。

## 2. 调度器

详细模型与调度流程参见 [Scheduler](scheduler/Scheduler.md)。

Go 调度器采用 GMP 模型：G 表示 goroutine，M 表示操作系统线程，P 表示执行 Go 代码所需的调度资源。M 持有 P 时可以执行 G。

核心调度流程位于 `proc.go`，可以从以下函数入手：

- `schedule()`：进入调度循环，选择后续执行的 goroutine。
- `findRunnable()`：寻找可运行的 goroutine，包括检查运行队列和工作窃取等。
- `execute()`：切换到选中的 goroutine 执行。

抢占相关逻辑主要位于 `preempt.go`，G、M、P 的结构定义主要位于 `runtime2.go`。

## 3. 垃圾回收

详细原理与调优方法参见 [GC](gc/GC.md)。

GC 的整体流程主要位于 `mgc.go`，可从 `gcStart()` 阅读启动条件与阶段切换。各职责进一步拆分为：

- `mgcmark.go`：对象扫描、标记及辅助标记。
- `mgcsweep.go`：清扫，回收不可达对象占用的空间。
- `mgcpacer.go`：控制 GC 触发时机与标记工作节奏。
- `mbarrier.go`：写屏障。

GC 通过并发标记清扫回收不可达的堆对象，部分阶段仍需要短暂的 STW 暂停。回收对象空间与向操作系统归还物理内存是不同过程，后者主要由 scavenger 负责。

## 4. 网络轮询器

工作流程、平台实现及调度器协作方式参见 [netpoller](netpoller.md)。

网络轮询器通过操作系统的 I/O 事件机制唤醒等待中的 goroutine，让多个连接可以共享线程等待事件。核心逻辑位于 `netpoll.go`，平台适配位于 `netpoll_*.go`。

## 5. 系统监控

监控线程的职责、执行方式与检查节奏参见 [sysmon](sysmon.md)。

sysmon 在不持有 P 的专门 M 上执行监控循环，检查抢占、系统调用与网络事件等状态，并在满足条件时唤醒 GC 等后台工作。核心实现位于 `proc.go`。

## 6. 阅读入口

- [runtime 源码](https://github.com/golang/go/tree/master/src/runtime)：查看运行时实现，阅读特定版本时切换对应 tag。
- [runtime 包文档](https://pkg.go.dev/runtime)：公开接口与运行时配置。
- [Go GC Guide](https://go.dev/doc/gc-guide)：GC 原理与调优。

[返回 Go 目录](../Go.md)
