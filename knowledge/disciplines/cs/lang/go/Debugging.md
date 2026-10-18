# Go 调试

Go 程序调试可以结合断点、日志、goroutine 栈、竞态检测、性能剖析和执行跟踪。先确定问题属于逻辑错误、并发等待还是性能瓶颈，再选择工具，通常比一开始收集全部数据更有效。

## 1. 工具选择

| 问题 | 常用工具 | 主要观察内容 |
| --- | --- | --- |
| 逻辑错误、变量值异常 | Delve、IDE 调试器 | 断点、调用栈、变量和执行路径 |
| goroutine 卡住或疑似泄漏 | goroutine 栈、goroutine profile | 等待位置、数量及重复堆栈 |
| 并发读写异常 | `go test -race` | 实际执行路径中的数据竞争 |
| CPU 或内存开销较高 | pprof | 热点调用栈、分配量与存活对象 |
| 调度延迟、阻塞和 GC 时序 | `go tool trace` | G、P、系统调用及 GC 等事件的时间关系 |
| 调度器或 GC 的整体状态 | `GODEBUG` | 运行队列、线程与 GC 摘要 |

## 2. 使用 Delve 断点调试

Delve 是面向 Go 的调试器，可直接使用命令行，也可以通过支持它的 IDE 操作。

```sh
# 调试当前 main 包
dlv debug .

# 调试当前包的测试
dlv test .
```

常用交互命令：

| 命令 | 用途 |
| --- | --- |
| `break main.main` | 在函数入口设置断点 |
| `continue` | 继续执行到断点或结束 |
| `next` | 单步执行，不进入普通函数调用 |
| `step` | 单步执行，尝试进入函数 |
| `print name` | 查看表达式或变量 |
| `stack` | 查看当前调用栈 |
| `goroutines` | 查看 goroutine 列表 |

优化和内联可能影响变量可见性与单步体验。需要单独构建调试二进制时，可以关闭优化和内联：

```sh
go build -gcflags='all=-N -l' -o /tmp/go-debug-app .
dlv exec /tmp/go-debug-app
```

关闭优化会改变性能和时序，性能测量应使用符合实际运行配置的构建。

## 3. goroutine 栈与等待问题

程序中可以通过 `runtime.Stack` 获取 goroutine 栈，或者使用 `runtime/pprof` 的 goroutine profile 输出堆栈。

```go
pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
```

这里的 `pprof` 来自 `runtime/pprof`。分析时重点看：

- 大量 G 是否长期停在相同的 channel、锁或网络等待位置。
- 等待是否符合业务预期，是否缺少关闭、取消或超时路径。
- 多次采样中，G 的数量和相同堆栈是否持续增长。

单次看到 G 处于等待状态不能直接判断泄漏。Unix 平台普通 Go 程序默认收到 `SIGQUIT` 时通常会打印堆栈并退出，因此它不适合作为保持进程运行的常规采样方式。

## 4. 数据竞争检测

```sh
go test -race ./...
go run -race .
```

竞态检测器报告缺少必要同步的并发内存访问。它只能检测实际执行到的路径；一次测试没有报告，不代表所有并发行为都正确。

该工具会增加运行开销，且有平台支持范围。死锁、goroutine 泄漏和业务执行顺序错误也不一定属于数据竞争。

## 5. pprof 性能剖析

测试和基准测试可以输出 CPU、内存 profile：

```sh
go test -run='^$' -bench=. -cpuprofile=cpu.out -memprofile=mem.out .
go tool pprof cpu.out
go tool pprof mem.out
```

这些命令会执行当前包的基准测试，应在有相关基准的包中使用。服务程序也可以通过 `runtime/pprof` 采集，或按需配置 `net/http/pprof`。

| 数据 | 适合分析 |
| --- | --- |
| CPU profile | CPU 时间主要消耗在哪些调用栈 |
| Heap profile | 堆对象存活量或分配开销，取决于所选采样指标 |
| Block profile | 同步操作上的阻塞，需要开启相应采样 |
| Mutex profile | 锁竞争，需要开启相应采样 |

CPU profile 不直接表示 G 等待网络或锁的总时长；分析等待时应结合对应 profile 或 trace。

## 6. 执行跟踪

```sh
go test -trace=trace.out .
go tool trace trace.out
```

程序中也可以通过 `runtime/trace` 主动采集。trace 适合观察 G 的创建、运行、等待、唤醒，以及 P、系统调用和 GC 等事件的时间关系。

pprof 更适合汇总“开销集中在哪里”，trace 更适合分析“这些事件如何按时间发生”。采集窗口应覆盖待排查场景，并控制数据量和采集开销。

## 7. 调度与 GC 摘要

```sh
GODEBUG=schedtrace=1000,scheddetail=1 ./app
GODEBUG=gctrace=1 ./app
```

`schedtrace` 按毫秒配置输出间隔，`scheddetail` 增加调度详情；`gctrace` 输出 GC 摘要。字段和支持选项随版本变化，应结合实际 Go 版本解释。

线程数量高不一定代表 CPU 并行度高，等待中的 G 多也不一定代表异常。需要把运行队列、P 使用情况、线程阻塞和业务负载结合起来判断。
