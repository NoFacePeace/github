# 异步提交远程执行

## 1. 处理内容

调用 SessionRunner，传入 Task、新 Session、最终配置、版本源资料、用户身份及适用的引擎会话恢复标识。调用以异步方式推进，接口返回 Task UUID、Session UUID 和当前状态。

远程分支后续准备内部工具上下文、执行输入、回调与调度请求。这些共用能力可从[准备内部工具上下文](../start-task/server/submit/PrepareInternalToolContext.md)沿后续链接阅读；无需重走创建 Task 和首轮 Session。

## 2. 下一步

进入[调度本轮执行](ScheduleExecution.md)。
