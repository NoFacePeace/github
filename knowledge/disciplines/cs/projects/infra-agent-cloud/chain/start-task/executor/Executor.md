# Executor

Runtime 在沙箱中启动 Executor 后，由 Executor 读取执行输入、适配具体引擎配置并启动 Agent 引擎。

## 1. 初始化流程

从[读取执行输入](read-execution-input/ReadExecutionInput.md)开始：

```text
读取并校验 manifest
  → 读取敏感输入并校验执行标识
  → 合成最终配置
  → 准备 MCP Gateway
  → 写入引擎配置
  → 构建引擎启动命令
  → 安装插件
  → 打开执行日志
  → 启动引擎进程
  → 持续读取并解析输出（与等待退出并行）
  → 等待引擎退出
  → 解析最终结果
  → 上报终态事件
  → 关闭执行资源
```

## 2. 职责边界

Runtime 管理外层隔离与进程生命周期；Executor 在已有环境内负责引擎适配，不重新进行 Server 的版本选择或调度。
