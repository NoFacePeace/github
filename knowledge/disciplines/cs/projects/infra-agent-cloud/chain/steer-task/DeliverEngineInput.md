# Executor 调用引擎追加能力

## 1. 处理内容

Executor 从控制命令流逐条取命令，按 commandId 查询本轮内存回执；首次处理时准备新提示词，调用当前引擎句柄的 steer，保存成功或失败结果。重复命令复用已有回执，避免再次写入引擎。

Claude 使用适用的流式交互，Codex 使用 app-server 交互；操作的是当前存活引擎进程，不是启动一个新进程恢复对话。具体共用能力见[处理追加输入](../start-task/executor/consume-engine-output/HandleControlInput.md)。

## 2. 后续流程

正常路径进入[回传追加指令回执](ReportSteerResult.md)。
