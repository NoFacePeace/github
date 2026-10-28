# 回传追加指令回执

## 1. 处理内容

Executor 发出 steer_result，携带 commandId 及成功结果或错误。成功结果可包含交互模式、是否已消费及适用的引擎会话、轮次标识。

回执写入 Executor 标准输出，经 Runtime 解析、校验后，沿原有[执行事件 Stream](../../storage/redis/RunnerEventStream.md)回传 Server，不另建专用结果 Stream。

## 2. 后续流程

正常路径进入[Server 保存控制回执](SaveControlReceipt.md)。
