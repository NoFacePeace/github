# Channel

channel 是 Go 在 goroutine 之间传递值并建立同步关系的机制。核心运行时实现位于 `src/runtime/chan.go`，编译器将创建、发送、接收和关闭等操作转换为相应的运行时调用或优化后的代码。

## 1. 内部结构

channel 的运行时结构为 `hchan`。以下是常见字段，具体布局以所用 Go 版本为准。

| 字段 | 作用 |
| --- | --- |
| `qcount` | 缓冲区当前元素数量 |
| `dataqsiz` | 缓冲区容量 |
| `buf` | 缓冲区地址 |
| `sendx`、`recvx` | 缓冲区发送、接收位置 |
| `sendq`、`recvq` | 等待发送、接收的队列 |
| `closed` | 关闭状态 |
| `lock` | 保护 channel 状态的锁 |

等待队列通过 `sudog` 等结构关联 G 和通信操作，不是调度器的运行队列。等待通信的 G 需要被唤醒后才能重新参与调度。

## 2. 发送与接收

### 2.1 无缓冲 channel

无缓冲 channel 的发送与接收需要配对。存在等待的另一方时，运行时可以直接传递数据并唤醒对方；没有对应操作时，当前 G 通常进入等待队列并挂起。

### 2.2 有缓冲 channel

有缓冲 channel 通常使用环形缓冲区。缓冲区有空间时，发送可以写入而不等待接收方；缓冲区有数据时，接收可以取出而不等待新的发送。

缓冲区满时发送可能阻塞，缓冲区空时接收可能阻塞。存在等待的通信方时，运行时还会通过直接传递或腾出缓冲空间等路径完成操作。

### 2.3 与调度器协作

阻塞通信通过 `gopark()` 等路径挂起 G，条件满足时通过 `goready()` 等路径将其置为可运行。挂起 G 不意味着执行它的 M 必须一直等待，M 可以执行其他可运行 G。

## 3. 关闭与 nil channel

| 操作 | 行为 |
| --- | --- |
| 向已关闭 channel 发送 | panic |
| 从已关闭 channel 接收 | 先取完缓冲数据，之后返回元素零值；双返回值形式的 `ok` 为 `false` |
| 重复关闭 channel | panic |
| 关闭 nil channel | panic |
| 向 nil channel 发送或从中接收 | 普通发送、接收操作永久阻塞 |

关闭 channel 会唤醒相关等待者；等待的发送者恢复后会因向已关闭 channel 发送而 panic。关闭不会立即丢弃缓冲区中的数据，也不是由 GC 自动完成的操作。

## 4. 同步与使用边界

channel 的发送和接收具有 Go 内存模型规定的同步关系，可用于传递数据与协调执行顺序。缓冲和无缓冲 channel 的具体同步规则有所区别，应以语言规范和内存模型为准。

多个 G 可以并发使用同一个 channel，但 channel 不能自动保护业务中的所有共享状态。也不能依赖多个并发发送者的启动顺序推断消息顺序。

## 5. 源码与参考资料

- `chan.go`：`hchan`、`makechan()`、`chansend()`、`chanrecv()`、`closechan()`。
- `runtime2.go`：`sudog` 等等待相关结构。
- [Channel types](https://go.dev/ref/spec#Channel_types)：channel 类型与语言语义。
- [Go Memory Model](https://go.dev/ref/mem)：同步关系。

## 6. 相关机制

- [Select](Select.md)：在多个 channel 通信操作之间选择。
- [Scheduler](scheduler/Scheduler.md)：通信阻塞后的等待与唤醒。

[返回 Runtime 目录](Runtime.md)
