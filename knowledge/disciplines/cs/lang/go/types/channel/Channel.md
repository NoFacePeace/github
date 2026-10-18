# Channel

channel 是 Go 在 goroutine 之间传递值并建立同步关系的机制。核心运行时实现位于 `src/runtime/chan.go`，编译器将创建、发送、接收和关闭等操作转换为相应的运行时调用或优化后的代码。

## 1. 设计理念

Go 对 channel 的设计理念常概括为“通过通信来共享内存”。它鼓励 goroutine 通过发送和接收数据来协作，将并发程序组织成明确的数据流与职责交接。

### 1.1 数据传递与执行协调

channel 不仅用于传递值，也用于协调执行进度。无缓冲 channel 的发送与接收必须配对，形成通信双方的交接点；有缓冲 channel 允许生产者与消费者在一定范围内独立推进，缓冲区满时发送方等待，从而形成背压。

例如，生产者提交任务，由消费者负责处理：

```go
jobs := make(chan Job)

go func() {
    for job := range jobs {
        process(job)
    }
}()

jobs <- job
close(jobs) // 此处假设只有一个发送方，且不再提交任务。
```

这里发送成功表示任务已被接收，不表示 `process(job)` 已执行完成。若需要等待处理完成，还应通过结果 channel、`sync.WaitGroup` 等机制表达完成条件。

### 1.2 数据所有权交接

使用 channel 时，可以约定：发送方交出数据后停止修改，由接收方继续负责。这样的职责划分有助于减少多个 goroutine 同时修改同一份数据的情况。

但 channel 不会自动转移或强制执行独占所有权。发送值遵循 Go 的赋值语义：发送指针、切片或 map 时，不会深拷贝其指向或关联的数据。如果双方仍并发访问这些数据，且至少一方执行写入，就需要额外同步以避免数据竞争。

### 1.3 适用场景

channel 适合表达任务交接、生产者与消费者、流水线、结果汇总和事件通知等通信流程。设计时应明确发送方、接收方、关闭责任，以及阻塞或取消时如何退出。

如果需求只是保护共享计数器、缓存或某段共享状态，使用互斥锁通常更直接。channel 与锁应根据协作方式选择，也可以在同一个程序中配合使用。

## 2. 内部结构

channel 的运行时结构为 `hchan`，组织缓冲区、发送与接收等待队列、关闭状态及同步锁。

- [Channel 内部结构](struct.md)：核心字段、环形缓冲区、等待队列及状态保护。

## 3. 发送与接收

- [Channel 发送](send.md)：编译入口、直接传值、缓冲区写入及阻塞发送的挂起与唤醒。
- [Channel 接收](recv.md)：编译入口、配对接收、缓冲区读取及关闭与阻塞接收的处理。

### 3.1 无缓冲 channel

无缓冲 channel 的发送与接收需要配对。存在等待的另一方时，运行时可以直接传递数据并唤醒对方；没有对应操作时，当前 G 通常进入等待队列并挂起。

### 3.2 有缓冲 channel

有缓冲 channel 通常使用环形缓冲区。缓冲区有空间时，发送可以写入而不等待接收方；缓冲区有数据时，接收可以取出而不等待新的发送。

缓冲区满时发送可能阻塞，缓冲区空时接收可能阻塞。存在等待的通信方时，运行时还会通过直接传递或腾出缓冲空间等路径完成操作。

### 3.3 与调度器协作

阻塞通信通过 `gopark()` 等路径挂起 G，条件满足时通过 `goready()` 等路径将其置为可运行。挂起 G 不意味着执行它的 M 必须一直等待，M 可以执行其他可运行 G。

## 4. 关闭与 nil channel

| 操作 | 行为 |
| --- | --- |
| 向已关闭 channel 发送 | panic |
| 从已关闭 channel 接收 | 先取完缓冲数据，之后返回元素零值；双返回值形式的 `ok` 为 `false` |
| 重复关闭 channel | panic |
| 关闭 nil channel | panic |
| 向 nil channel 发送或从中接收 | 普通发送、接收操作永久阻塞 |

关闭 channel 会唤醒相关等待者；等待的发送者恢复后会因向已关闭 channel 发送而 panic。关闭不会立即丢弃缓冲区中的数据，也不是由 GC 自动完成的操作。

## 5. 同步与使用边界

channel 的发送和接收具有 Go 内存模型规定的同步关系，可用于传递数据与协调执行顺序。缓冲和无缓冲 channel 的具体同步规则有所区别，应以语言规范和内存模型为准。

多个 G 可以并发使用同一个 channel，但 channel 不能自动保护业务中的所有共享状态。也不能依赖多个并发发送者的启动顺序推断消息顺序。

## 6. 源码与参考资料

- `chan.go`：`hchan`、`makechan()`、`chansend()`、`chanrecv()`、`closechan()`。
- `runtime2.go`：`sudog` 等等待相关结构。
- [Effective Go：Share by communicating](https://go.dev/doc/effective_go#sharing)：通过通信共享内存的设计理念。
- [Channel types](https://go.dev/ref/spec#Channel_types)：channel 类型与语言语义。
- [Go Memory Model](https://go.dev/ref/mem)：同步关系。

## 7. 相关机制

- [Select](../../keywords/Select.md)：在多个 channel 通信操作之间选择。
