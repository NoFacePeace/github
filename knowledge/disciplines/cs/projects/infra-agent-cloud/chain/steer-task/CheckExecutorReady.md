# 检查租约与执行器就绪

## 1. 处理内容

远程分支读取[执行租约](../../storage/redis/ExecutionLease.md)，确认归属当前 Task；再读取本轮 executor_ready 控制回执，并校验它与当前租约一致。

就绪回执按通用格式存于：

```text
<namespace>:<scope>:runner:events:{<taskUuid>:<sessionUuid>}:control:ready
```

没有匹配租约或就绪回执时，不提供当前执行的投递能力。随后仍会登记适用的排队回执，不能仅据此断定消息永久丢弃。

## 2. 后续流程

正常路径进入[认领回执并保存指令](ClaimDeliveryReceipt.md)。
