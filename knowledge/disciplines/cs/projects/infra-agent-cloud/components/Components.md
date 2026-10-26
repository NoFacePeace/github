# 组件

整理 Agent 云平台各组件的职责与协作关系。

## 1. 服务端组件

- [Server](Server.md)：处理业务请求、授权、任务管理与调度。
- [Runtime](Runtime.md)：准备执行环境，管理任务执行进程的生命周期。
- [Executor](Executor.md)：适配 Agent 引擎，负责启动、配置与输出解析。

Server 发起并调度执行，Runtime 承载执行环境并启动 Executor，Executor 调用具体的 Agent 引擎。执行事件与结果沿相应通道回传。
