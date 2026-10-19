# Perplexity SPACE

SPACE（Sandboxed Platform for Agentic Code Execution）是 Perplexity 开源的沙箱平台，用于在隔离的 Firecracker microVM 中运行 Agent 和用户代码，采用 Apache 2.0 许可证。

## 1. 核心能力

- 创建执行环境：从模板创建带有独立虚拟机和可写文件系统的 Linux 沙箱。
- 执行命令：通过 HTTP API 创建命令会话、启动进程并读取输出与退出结果。
- 管理生命周期：保留、暂停、挂起、恢复和删除沙箱。
- 配置运行设施：提供网络出口策略、可观测性、预热池和基础快照等相关文档。

## 2. 与 Agent 沙箱的关系

SPACE 是 [Agent 沙箱](../../agent/sandbox/Sandbox.md) 的具体实现，可为命令执行、文件操作和代码运行提供隔离环境。

应用决定执行什么任务、如何调用模型和保留哪些结果；SPACE 管理沙箱、进程和命令输出连接。独立执行环境的实际访问范围仍取决于挂载、网络和权限配置。

## 3. 使用流程

1. 按官方安装指南部署并配置认证；入门教程使用本地 Kind 集群。
2. 从模板创建沙箱，等待环境就绪。
3. 创建命令会话并连接，启动进程。
4. 读取输出和退出结果，收集需要保留的产物。
5. 根据后续任务需要保留、暂停或删除沙箱。

沙箱生命周期操作是异步的，API 返回操作标识，调用方通过查询获取完成状态。SPACE 以 Rust 实现，应用通过 HTTP API 使用，无需嵌入 Rust 库。

## 4. 参考资料

- [官方 GitHub 仓库与许可证](https://github.com/perplexityai/SPACE)
- [SPACE Guide](https://space.docs.perplexity.ai/guide/index.html)
- [第一个沙箱教程](https://space.docs.perplexity.ai/guide/first-sandbox.html)
