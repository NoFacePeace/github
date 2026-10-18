# waiting

`waiting` 是 Go runtime 的 `g` 结构体字段，定义于 `src/runtime/runtime2.go`，类型为 `*sudog`，用于关联这个 G 的相关等待记录。节点结构与用途参见 [sudog](../../sudog.md)。

## 1. 字段与链表

```go
// 简化示意，并非完整结构定义。
type g struct {
    waiting *sudog
}
```

`waiting` 指向等待记录链的首节点，后续节点通过 `sudog.waitlink` 串联：

```text
g.waiting → sudog A → sudog B → …
               waitlink   waitlink
```

它不是 channel 的等待队列头。节点可以同时通过 `next`、`prev` 加入 channel 队列，通过 `waitlink` 加入当前 G 的等待记录链，两种链接承担不同职责。

## 2. 普通 channel 等待

普通阻塞发送或接收通常只有一个等待节点。运行时登记 `sudog` 并关联到 `g.waiting`，再通过挂起协议暂停当前 G。

通信或关闭操作唤醒 G 后，恢复路径清理等待信息并归还节点。等待期间的记录也用于运行时协调栈移动等操作。

## 3. select 中的多个节点

阻塞的 `select` 可以为多个非 nil channel 分支登记等待节点。这些节点分别位于各自 channel 的发送或接收队列，同时通过 `waitlink` 串成当前 G 的等待记录链。

一个分支选中并唤醒 G 后，`selectgo` 在持有相关 channel 锁的情况下遍历本次等待记录，清理未选中分支仍在队列中的节点，再释放记录。

从 channel 队列移除节点时，可以利用 `next`、`prev` 直接摘除，无需在每个 channel 队列中重新搜索整个 G。具体遍历及分支对应关系以所用版本的 `selectgo` 实现为准。

## 4. 使用边界

`waiting` 不代表所有等待机制的统一队列。GC worker、计时器及其他运行时机制可以通过不同结构关联等待的 G，不能仅凭此字段是否为空判断 G 是否处于等待状态。

该字段与节点生命周期由运行时维护，不是业务代码可直接访问或修改的公开 API。
