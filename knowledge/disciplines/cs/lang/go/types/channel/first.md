# first

`first` 是 `waitq` 的队首指针，类型为 `*sudog`，指向队列中的第一个等待节点。

## 1. 状态含义

- 空队列中，`first` 与 `last` 均为 `nil`。
- 只有一个节点时，`first` 与 `last` 指向同一节点。
- 多个节点时，从 `first` 沿节点的 `next` 字段可以依次访问后续节点。

队首节点的 `prev` 为 `nil`。

## 2. 队列操作中的变化

向空队列入队时，`first` 与 `last` 一起指向新节点。向非空队列尾部追加节点时，`first` 保持不变。

从队首移除节点时，`first` 更新为原队首的 `next`。如果还有节点，新队首的 `prev` 被设为 `nil`；如果队列已空，`last` 也被设为 `nil`。

`dequeue` 会跳过已由其他 `select` 分支赢得选择的等待节点，所以原 `first` 指向的节点不一定是最终返回的节点。相关操作由调用方通过 channel 的锁保护。

## 3. 相关结构

- [sudog](../../runtime/sudog.md)：队首指针指向的等待节点，关联 goroutine 与等待操作。
