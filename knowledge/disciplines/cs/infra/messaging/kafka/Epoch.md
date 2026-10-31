---
color: "#5B6CDA"
---

# Epoch

Kafka 中的 epoch 通常表示特定对象的代次或任期，用于识别过期状态、隔离旧实例，并辅助日志一致性检查。这里的 epoch 不是 Unix 计时起点，也不是时间戳。

## 1. 作用域

Kafka 没有一个适用于所有对象的统一 epoch。不同类型的 epoch 绑定不同对象，必须先确定对象，再比较它的代次。

| 类型 | 绑定对象 | 含义 |
| --- | --- | --- |
| Leader Epoch | 具体分区。 | 该分区 Leader 的任期编号。 |
| Producer Epoch | Producer ID，简称 PID。 | 该生产者身份的代次。 |
| Partition Epoch | KRaft 中的具体分区。 | 该分区元数据的变更版本。 |

不同分区的 Leader Epoch 没有直接比较新旧的意义；Leader Epoch 和 Producer Epoch 也不能相互比较。

## 2. Leader Epoch

### 2.1 与分区和 Broker 的关系

Leader Epoch 绑定分区，不是 Broker。同一个 Broker 可以担任多个分区的 Leader，各分区独立维护自己的 Leader Epoch。

```text
orders-0：LeaderId = 1，LeaderEpoch = 7
orders-1：LeaderId = 1，LeaderEpoch = 3
```

两个分区的 Leader 都是 Broker 1，但它们的 epoch 不必相同。如果只有 `orders-0` 发生 Leader 切换，其任期变化不会直接导致 `orders-1` 的 epoch 变化。

同一个 Broker 再次担任同一分区的 Leader，也属于新的任期：

```text
orders-0：
  Broker 1 担任 Leader：LeaderEpoch = 7
  切换到 Broker 2：     LeaderEpoch = 8
  再切回 Broker 1：     LeaderEpoch = 9
```

只比较 Broker ID，无法区分 Broker 1 的前后两次任期；结合 Leader Epoch 才能区分。

### 2.2 分区元数据

Leader Epoch 存在于分区元数据中。`MetadataResponse` 从协议版本 v7 开始，在每个分区的信息中返回 `LeaderEpoch`；这里的 v7 是该 API 的协议版本，不是 Kafka 产品版本。

| 字段 | 含义 |
| --- | --- |
| `PartitionIndex` | Topic 内的分区编号。 |
| `LeaderId` | 当前担任该分区 Leader 的 Broker ID，不是另外生成的 Leader 身份编号。 |
| `LeaderEpoch` | 该分区当前的 Leader 任期编号。 |
| `ReplicaNodes` | 承载该分区副本的 Broker ID 列表。 |
| `IsrNodes` | 与 Leader 保持同步的副本所在的 Broker ID 列表。 |

响应中还包含 Broker 列表，通过其中的 `NodeId`、`Host` 和 `Port`，客户端可以把 `LeaderId` 映射到对应的 Broker 地址。

### 2.3 过期请求与日志恢复

对于携带有效当前 Leader Epoch 的 Fetch 等支持 epoch 校验的请求，Broker 会按目标分区进行比较：

- 请求中的 epoch 小于 Broker 已知的 epoch：请求基于旧任期，返回 `FENCED_LEADER_EPOCH`。
- 请求中的 epoch 大于 Broker 已知的 epoch：Broker 尚不知道该任期，返回 `UNKNOWN_LEADER_EPOCH`。
- 两者一致：通过 epoch 校验，但请求仍需满足其他检查。

例如，`orders-0` 已进入 epoch 8，携带 epoch 7 的 Fetch 请求就能被识别为过期。不能把 `orders-1` 的 epoch 拿来校验这个请求。

日志恢复时，Kafka 还利用 Leader Epoch 与日志 offset 的对应关系，帮助副本定位日志分歧并确定截断位置。offset 表示日志位置，Leader Epoch 用于区分日志所处的 Leader 任期；epoch 本身不是完整的恢复协议。

## 3. Producer Epoch

Producer Epoch 与 PID 配合，标识生产者身份的某个代次，而不是某个分区的 Leader 任期。

```text
旧代次：(PID = 100，ProducerEpoch = 2)
新代次：(PID = 100，ProducerEpoch = 3)
```

在事务场景中，`transactional.id` 提供跨会话的稳定逻辑身份，事务协调器维护它对应的 PID 和 Producer Epoch。新实例使用相同的 `transactional.id` 初始化并接管时，可以通过更新生产者身份的代次隔离旧实例，避免旧进程恢复后继续操作事务。

Producer Epoch 不是网络连接编号，普通网络重连不意味着它一定递增。它也不是独立完成幂等写入的机制；幂等写入还需要 PID、按分区维护的消息序列号以及 Broker 侧状态校验。

## 4. Partition Epoch

KRaft 的内部 `PartitionRecord` 同时包含 `LeaderEpoch` 和 `PartitionEpoch`：

- `LeaderEpoch`：描述分区 Leader 的任期。
- `PartitionEpoch`：描述分区元数据的变更版本，分区状态变更时递增。

例如，ISR 变化不一定伴随 Leader 切换，因此不能把分区元数据的每次变化都理解为一次新的 Leader 任期。

## 5. 参考资料

- [MetadataResponse 协议定义](https://github.com/apache/kafka/blob/trunk/clients/src/main/resources/common/message/MetadataResponse.json)。
- [KIP-320：Fetch 请求校验与日志截断检测](https://cwiki.apache.org/confluence/spaces/KAFKA/pages/87295403/KIP-320+Allow+fetchers+to+detect+and+handle+log+truncation)。
- [Kafka 协议：InitProducerId 与 Producer Epoch](https://kafka.apache.org/protocol.html)。
- [KafkaProducer API：事务生产者](https://kafka.apache.org/41/javadoc/org/apache/kafka/clients/producer/KafkaProducer.html)。
- [PartitionRecord 元数据定义](https://github.com/apache/kafka/blob/trunk/metadata/src/main/resources/common/metadata/PartitionRecord.json)。
