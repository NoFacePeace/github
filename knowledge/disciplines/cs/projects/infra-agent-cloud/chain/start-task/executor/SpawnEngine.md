# 启动引擎进程

Executor 使用前面构建的命令调用 Node.js `spawn`，创建 Claude Code 或 Codex 引擎子进程。

## 1. 启动设置

- 工作目录使用 Executor 当前工作目录。
- 环境变量合并当前进程环境和引擎专用环境。
- 标准输出与标准错误使用管道，交给 Executor 消费。
- 支持执行中追加输入的模式使用标准输入管道，普通模式忽略标准输入。

等待进程启动成功后，立即登记退出监听，避免短任务在初始化期间结束而漏掉退出事件。

## 2. 追加输入模式

该模式还会初始化 Codex app-server 客户端或 Claude 流式交互，将提示词、工作目录和适用的恢复会话标识交给引擎。

引擎进程由 Executor 创建；外层 bwrap 沙箱和 Executor 进程此前已经由 Runtime 启动。

## 3. 存储操作

创建进程本身不直接读写平台 MySQL、Redis 或对象存储。引擎运行时会读取配置，并按自身行为访问任务文件。

## 4. 下一步

启动后[持续读取并解析输出](consume-engine-output/ConsumeEngineOutput.md)。
