# waitq

`waitq` 是 Go runtime 为 channel 定义的等待队列类型，位于 `src/runtime/chan.go`，用于组织等待发送或接收的 `sudog` 节点。

## 1. 结构定义

```go
type waitq struct {
    first *sudog
    last  *sudog
}
```

`first` 指向队首，`last` 指向队尾；空队列中两者均为 `nil`。节点之间通过 `sudog` 的 `next`、`prev` 字段连接，组成双向链表。

- [first](first.md)：队首指针的状态及入队、出队时的变化。
- [last](last.md)：队尾指针的状态及追加、移除节点时的变化。

## 2. 在 channel 中的用途

`hchan` 包含两个 `waitq` 字段：

| 字段 | 用途 |
| --- | --- |
| `sendq` | 保存等待发送的节点 |
| `recvq` | 保存等待接收的节点 |

普通发送或接收无法立即完成且允许阻塞时，运行时将相应节点加入队列并挂起 goroutine。通信条件满足或 channel 关闭时，运行时处理相关等待节点并唤醒 goroutine。

队列保存的是等待通信的记录，缓冲数据由 `hchan.buf` 保存。

## 3. 入队与出队

`enqueue` 将节点追加到队尾，`dequeue` 从队首取出可用节点。相关队列操作由调用方通过 channel 的锁保护，`waitq` 自身不包含锁。

对于 `select`，一个 goroutine 可以在多个 channel 上登记等待节点。出队时需要协调选择结果，跳过已经由其他分支赢得选择的节点。因此，不能仅凭队列顺序推断 goroutine 最终执行或完成的顺序。

## 4. 与 sudog 的区别

`waitq` 是 channel 使用的具体队列类型；等待队列这一概念并非 channel 专有。

`sudog` 是更通用的运行时等待节点，定义于 `src/runtime/runtime2.go`，除 channel 外，运行时的信号量等待机制等也会使用。它关联 goroutine 与等待操作，同一个 goroutine 可以对应多个等待节点。
