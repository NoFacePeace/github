# 发布追加指令命令

## 1. 处理内容

生成 commandId，读取并校验本轮租约，再通过 Runner 心跳能力标记确认支持 Steer。

通过 Redis 控制命令通道发布 steer 命令，携带 Task、Session、Runner、范围、租约及 prompt、messageId、幂等标识。命令有效期取租约截止时间与当前时间加 30 秒中较早者。

发布命令后，Server 开始等待对应控制回执；等待与 Runner 后续处理同时进行，不代表发布成功即完成引擎投递。

## 2. 后续流程

正常路径进入[Runner 校验控制命令](ValidateRunnerCommand.md)。
