# OpenCode

OpenCode 是由 Anomaly 开发维护的开源 AI 编程 Agent，项目仓库为 `anomalyco/opencode`。它围绕模型和工具推进任务，通过文件操作、命令执行及外部工具完成开发工作。

## 1. 扩展机制

- [插件](plugins/Plugins.md)：通过事件钩子、自定义工具等方式扩展运行行为。
- MCP：连接外部工具与数据服务。
- Skills：提供可复用的任务说明与配套资源。

插件属于运行框架的扩展机制，模型是否参与取决于具体能力：工具通常由模型调用，事件钩子则在对应阶段由宿主触发。

## 2. 参考资料

核对日期：2026-10-21。

- [OpenCode 仓库](https://github.com/anomalyco/opencode)
- [OpenCode 文档](https://opencode.ai/docs/)
- [插件文档](https://opencode.ai/docs/plugins/)
