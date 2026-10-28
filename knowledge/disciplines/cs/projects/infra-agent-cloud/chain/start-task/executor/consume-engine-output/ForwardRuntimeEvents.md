# Runtime 接收并转发事件

Runtime 逐行读取 Executor 的标准输出，将 JSON 解析为协议事件后，校验并交给事件发送通道。

## 1. 协议校验

- 校验事件结构。
- 确认 `executionId` 与当前执行一致。
- 确认 `seq` 从 1 开始连续递增。

协议错误会触发停止执行。非终态事件持续转发；终态事件先保留，外层完成适用的收尾后再发送。

## 2. Redis 写入

Runtime 的事件发布通道通过 `XADD` 写入[执行事件 Stream](../../../../storage/redis/RunnerEventStream.md)，并维护[事件积压字节计数](../../../../storage/redis/RunnerEventBytes.md)。

通用 key 为：

```text
<namespace>:<replyTo>:runner:events
```

Stream 的 `payload` 字段保存事件 JSON，包含任务、Session、Runner、租约及适用的执行标识。Server 消费后继续处理业务状态和回复。

Runtime 按顺序提交输出事件，最多积累 32 个待完成的处理 Promise 后等待这一批完成；结束时也会等待全部发送完成。该窗口限制与 Redis 队列容量限制是两层控制。

## 3. 后续流程

这条输出链随引擎输出持续运行，同时 Executor[等待引擎退出](../WaitForEngineExit.md)。退出后等待剩余输出处理完成，再进入结果解析。
