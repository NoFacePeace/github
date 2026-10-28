# 向 Executor 输入管道写入

## 1. 处理内容

Runtime 查找当前 Session 对应的活动执行，校验 Task 与租约一致、执行尚未收尾且句柄支持 sendInput。

将 payload 校验为追加指令协议，补充 executionId 和 commandId，序列化为一行 JSON，写入 Executor 的输入管道。该步骤是进程间通信，不直接写 Session 表。

执行不存在、不匹配或已收尾时不写入。控制命令被 Runner 消费不代表引擎已接受，也不能把传输确认当作成功回执。

## 2. 后续流程

正常路径进入[Executor 调用引擎追加能力](DeliverEngineInput.md)。
