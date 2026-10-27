# 准备内部工具上下文

首个 Session 创建后，检查本轮是否需要由 Server 提供的内部工具。需要时，准备工具运行资料和访问校验信息，供 Agent 后续通过 MCP 调用。

## 1. 触发条件

根据所选模型等运行配置，确定需要注入的内部工具。没有对应工具时，跳过上下文创建，继续后续流程。

此时尚未调用工具或模型服务，只准备调用所需的配置。

## 2. 创建与保存

生成上下文 ID 和随机访问令牌，将令牌哈希及各工具的运行资料保存为 JSON，写入 Redis 的[内部工具上下文](../../../../storage/redis/InternalToolContext.md)。上下文初始有效期为 2 小时，验证访问时续期；Redis 不可用时，当前实现可回退到进程内存。

以下是图片分析工具的通用示例，字段值均为虚构：

```json
{
  "tokenHash": "访问令牌的哈希",
  "toolContexts": {
    "image_analysis": {
      "apiKey": "示例模型服务密钥",
      "baseUrl": "https://model.example.com",
      "workDir": "/workspace/task-example",
      "allowedFiles": ["/uploads/chart.png"]
    }
  }
}
```

上下文包含实际工具配置，可能含有敏感值；哈希处理仅针对访问令牌，不表示其他字段已经加密。示例工具名仅用于说明，不绑定具体内部实现。

## 3. 注入工具连接

将包含上下文 ID 的 MCP 地址、访问令牌和允许使用的工具列表加入本轮工具配置。必要时追加工具使用说明。

Agent 执行中调用工具时，Server 根据上下文 ID 查找资料，验证令牌，再将对应工具资料与调用参数交给工具处理。

## 4. 生命周期

Redis 保存上下文本体；Session 后续仅保存上下文 ID，供执行结束或启动失败时撤销。此阶段写入 Redis，与构建执行输入后更新 Session 是两个操作。

回退到进程内存时，上下文仅在当前进程可用，不能将其理解为跨实例共享存储。

## 5. 后续步骤

进入[构建远程执行输入](build-remote-execution-input/BuildRemoteExecutionInput.md)，将已注入工具配置与其他配置组织为远程执行输入。输入构建完成后，再按需将上下文 ID 更新到 Session，随后登记执行回调并提交调度。
