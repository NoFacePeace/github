# Server 保存控制回执

## 1. 处理内容

Server 消费并校验执行事件，在保存事件的 Redis Lua 流程中，为 steer_result 按 commandId 保存独立的控制回执：

```text
<namespace>:<scope>:runner:events:{<taskUuid>:<sessionUuid>}:control:<commandId>
```

该 key 使用 String 保存事件 JSON，并有过期时间及首次写入保护。它和 SSE 的事件回放 Sorted Set、日志 String 是不同 key；控制回执的保存不依赖是否保留供 SSE 消费的回放成员。

共用消费流程见 [Server 事件处理](../../components/server/handle-events/HandleEvents.md)。

## 2. 后续流程

正常路径进入[等待回执并确认投递](ConfirmDelivery.md)。
