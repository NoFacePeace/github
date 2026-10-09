---
color: "#5B6CDA"
---

# Fencing

Fencing 是隔离旧写入者的机制。在 Pulsar 中，需要区分底层 BookKeeper 的 ledger fencing 与 Producer 的 fencing。

## 1. 术语

- **Fencing**：机制或过程，适合作为文档标题，例如 ledger fencing。
- **Fence**：动词，表示执行隔离动作。
- **Fenced**：表示已经被隔离的状态，例如 BookKeeper 的 `LedgerFencedException`。
- **ExclusiveWithFencing**：Pulsar Producer 访问模式的正式名称。

## 2. Ledger Fencing

Pulsar Broker 将 Topic 的持久化消息写入 BookKeeper 的 ledger。Broker 故障或 Topic 所有权转移后，新 Broker 在恢复旧 ledger 时会通过 BookKeeper 执行 fencing，阻止旧写入者继续成功追加数据。

恢复流程确定旧 ledger 的最后有效记录并将其关闭，随后 Pulsar 可以创建新的 ledger 继续写入。即使旧 Broker 恢复运行，也不能继续成功写入已经被 fence 的 ledger，从而避免新旧写入者并发写入同一个 ledger。

Fencing 本身并不等同于数据恢复完成，也不会删除已有消息。

## 3. Producer Fencing

Pulsar 的 `ExclusiveWithFencing` 访问模式允许新 Producer 获取 Topic 的独占发布权，并隔离已有 Producer。被隔离的 Producer 无法继续在该 Topic 上发布消息。

这与 ledger fencing 发生在不同层面：前者控制客户端 Producer 的发布权，后者控制对 BookKeeper ledger 的写入。

## 4. 区别

| 机制 | 隔离对象 | 作用范围 |
| --- | --- | --- |
| Ledger fencing | 旧 ledger 写入者，通常是 Broker | BookKeeper ledger |
| Producer fencing | 已有 Producer | Topic 的发布权 |

这里的 fencing 是写入权限隔离机制，与 CPU 的内存屏障无关。
