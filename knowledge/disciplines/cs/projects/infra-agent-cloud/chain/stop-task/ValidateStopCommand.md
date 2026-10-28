# 接收并校验停止命令

## 1. 处理内容

Runner 消费控制命令后，校验其适用身份与有效性，并确认 Session 的活动租约 ID 与命令一致。旧租约命令不作用于新的执行。

有效 stop 命令交给执行服务，使用 Session UUID 与 leaseId 定位当前执行。

## 2. 后续流程

[停止准备或运行中的执行](StopCurrentExecution.md)。
