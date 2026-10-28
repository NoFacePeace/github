# 启动引擎并恢复对话

## 1. 处理内容

Runtime 在沙箱中启动新的 Executor，Executor 读取输入、生成配置并启动新的 Agent 引擎进程。流程见[启动 Executor](../start-task/runtime/start-executor/StartExecutor.md)。

有适用的 `resumeSessionId` 时使用引擎恢复能力；没有时启动新的引擎会话。平台 Session 始终是本轮新建记录，恢复的是引擎对话上下文，不是旧 Session 的进程。

## 2. 下一步

进入[回传结果与执行收尾](HandleResult.md)。
