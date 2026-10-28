# 执行事件 Stream

保存 Runner 回传的接受、进度及执行结果等事件，供 Server 消费。

## 1. Key 与类型

以下使用通用命名空间：

```text
<namespace>:<replyTo>:runner:events
```

类型为 Stream。`replyTo` 是 Job 中的回复范围，不是 Runner ID。该流承载同一回复范围内多轮执行的事件，并非每个 Session 单独一个 key。

## 2. 内容与写入

通过 `XADD` 追加消息，字段 `payload` 保存事件 JSON，包含类型、任务和执行标识、Runner、租约及序号等信息。

Stream 消息 ID 与事件 JSON 内的 `messageId` 是不同标识。

## 3. 积压控制

发布时原子检查队列字节数与条目数，未超过限制才追加并更新字节计数。队列已满则等待消费者清理空间，等待超时后报告错误。

事件流与用于回调登记的执行事件状态 Hash 是不同 key，不能混为一谈。
