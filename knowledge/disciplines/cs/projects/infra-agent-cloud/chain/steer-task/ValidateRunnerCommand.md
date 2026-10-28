# Runner 校验控制命令

## 1. 处理内容

Runner 消费控制命令并按适用的目标、有效期和租约约束处理。命令租约与活动租约不匹配时忽略，防止旧命令作用于新的执行。

有效 steer 命令传给执行服务，参数包含 Task UUID、Session UUID、leaseId、commandId 和 payload。

## 2. 后续流程

正常路径进入[向 Executor 输入管道写入](SendExecutorInput.md)。
