# 持续回传执行输出

## 1. 处理内容

引擎接收新指令后，继续使用当前执行的输出管道与事件通道。Runtime 回传进度和输出，Server 保存适用日志、回放并通过已有回调推送 SSE。

共用流程见[持续读取并解析输出](../start-task/executor/consume-engine-output/ConsumeEngineOutput.md)。Steer 本身不为成功追加的消息立即创建新平台 Session，也不重新走 Task 创建、调度和沙箱启动。
