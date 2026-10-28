# 执行收尾并回传终态

## 1. 处理内容

Runtime 等待执行结束并收齐输出，停止本轮日志定时器、处理适用认证回写、清理环境，登记后台备份，再发布最终结果。具体职责见[执行收尾](../start-task/runtime/finalize-execution/FinalizeExecution.md)。

终态通过[执行事件 Stream](../../storage/redis/RunnerEventStream.md)送往 Server。异常路径可能回传 failed；Server 还会结合停止请求标记决定业务终态。

后台文件备份不要求在终态可见前全部完成。

## 2. 后续流程

[保存 Session 停止终态](PersistStoppedSession.md)。
