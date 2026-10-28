# 回传 accepted 事件

Job 校验与维护检查通过后，Runner 发布 `accepted` 事件，通知 Server 本轮任务已被接受。此时尚未启动 Agent。

## 1. 构建事件

组织事件类型、消息标识、时间、追踪信息，以及 Job、Task、Session、Runner、租约和事件序号等字段。回复范围取自 Job 的 `replyTo`，用于确定 Server 接收事件的范围。

适用时附带执行授权引用；事件本身不包含全部凭证明文。

## 2. 写入 Redis

向[执行事件 Stream](../../../storage/redis/RunnerEventStream.md)追加事件，同时更新[事件字节计数](../../../storage/redis/RunnerEventBytes.md)。两个 key 在同一 Lua 脚本中处理。

写入前检查事件大小、队列字节数和条目数。超过积压限制时异步等待，超过等待时限则报错，不无限等待。当前发送流程会暂停，但不阻塞整个进程的事件循环。

## 3. 后续步骤

事件成功发布后，进入[准备执行环境](prepare-execution-environment/PrepareExecutionEnvironment.md)，调用执行服务准备凭证、资源、工作目录与隔离环境。Server 异步消费事件，再按执行归属与状态处理业务记录；Runner 不在本步骤直接写 MySQL。

`accepted` 不等于 Job 消费确认 `XACK`，也不表示 Agent 已启动或执行完成。
