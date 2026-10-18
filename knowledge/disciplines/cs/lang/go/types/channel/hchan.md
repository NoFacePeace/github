# hchan

`hchan` 是 Go channel 的核心运行时结构，定义于 `src/runtime/chan.go`，用于组织缓冲区、发送与接收等待队列、关闭状态及同步锁。

## 1. 核心字段

以下列出理解通信流程时常用的字段，并非完整结构定义；字段和具体布局以所用 Go 版本为准。

| 字段 | 作用 |
| --- | --- |
| `qcount` | 缓冲区当前元素数量 |
| `dataqsiz` | 缓冲区容量 |
| `buf` | 缓冲区地址 |
| `elemsize` | 单个元素的大小 |
| `elemtype` | 元素类型信息 |
| `sendx` | 下一次向缓冲区写入的位置 |
| `recvx` | 下一次从缓冲区读取的位置 |
| `sendq` | 等待发送的队列 |
| `recvq` | 等待接收的队列 |
| `closed` | 关闭状态 |
| `lock` | 保护 channel 状态及相关等待节点的锁 |

## 2. 字段之间的关系

- `buf`、`dataqsiz`、`qcount`、`sendx` 和 `recvx` 共同描述环形缓冲区的存储及读写位置；无缓冲 channel 的 `dataqsiz` 为零。
- `sendq` 和 `recvq` 组织等待通信的 `sudog` 节点，关联等待的 goroutine 与通信操作。
- `closed` 记录关闭状态；关闭后仍可接收缓冲区中剩余的数据。
- `lock` 协调缓冲区、队列和关闭状态等并发修改，不负责保护业务中通过 channel 传递的共享对象。

## 3. 相关知识

- [waitq](waitq.md)：`sendq` 与 `recvq` 使用的等待队列类型。
- [Channel 内部结构](struct.md)：缓冲区、等待队列和状态保护的整体关系。
