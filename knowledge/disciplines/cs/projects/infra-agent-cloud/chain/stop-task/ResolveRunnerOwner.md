# 定位并校验 Runner

## 1. 处理内容

停止命令发布方法会再次尝试取消排队，处理调用间隙的状态变化；仍未取消时，查询并校验[任务与 Runner 映射](../../storage/mysql/session-runner-mapping/SessionRunnerMapping.md)。

要求映射有效、运行范围匹配且 Runner 心跳可用，并核对本轮租约中的 Task、Runner、范围及适用的租约标识。校验失败可以返回错误，不能声称命令已送达。

## 2. 后续流程

再次取消成功走[排队取消结果](FinalizeQueueCancellation.md)；校验通过后[发布停止命令](PublishStopCommand.md)。
