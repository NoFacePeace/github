# 发布停止命令

## 1. 处理内容

生成 stop 命令，携带命令与消息标识、Task UUID、Session UUID、目标 Runner、运行范围、租约 ID 和适用的过期时间，通过 Redis 命令传输通道发布。

命令通道用于 Server 向 Runner 发控制指令，与 Job 投递 Stream、Runner 回传事件 Stream 用途不同。发布成功表示命令已交给传输层，不表示引擎已退出。

## 2. 后续流程

Runner 异步[接收并校验停止命令](ValidateStopCommand.md)；Server 接口继续[更新 Task 状态](UpdateTaskStatus.md)，不等待本轮完整退出。
