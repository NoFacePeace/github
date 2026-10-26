# 内部工具上下文

保存一轮执行调用 Server 内部工具所需的访问校验信息与运行资料。

## 1. Key

以下使用通用命名空间，不记录内部项目的实际前缀：

```text
<namespace>:internal-mcp:<contextId>
```

`contextId` 是新生成的 UUID，用于唯一标识工具上下文；它与 Task、Session 的标识分别使用。

## 2. 类型与内容

Redis 类型为 String，值为序列化 JSON：

```json
{
  "tokenHash": "访问令牌的 SHA256 哈希",
  "toolContexts": {
    "example_tool": {
      "workDir": "/workspace/task-example"
    }
  }
}
```

- `tokenHash`：用于验证调用方携带的访问令牌，不保存该令牌的原文。
- `toolContexts`：以工具名称为键，保存对应工具需要的配置；具体结构由工具决定，可包含服务地址、凭证和允许访问的文件等信息。

示例仅表达结构。令牌使用哈希不表示工具资料中的其他敏感值也已加密。

## 3. 写入与读取

准备内部工具上下文时生成 ID 和令牌，并通过 `SET` 写入 JSON，同时设置有效期。添加工具资料时，先验证已有上下文，再合并并保存。

Agent 调用 MCP 时，通过连接地址传入上下文 ID，并在请求头携带令牌。Server 根据 ID 读取本 key，校验令牌后取出对应工具资料；无需先从 Session 查询 ID。

## 4. 有效期与清理

初始有效期为 2 小时，验证访问时续期。执行结束或启动失败时撤销上下文，删除对应 key。

Session 仅保存上下文 ID，供后续清理定位。Redis 不可用时可回退到进程内存，此时数据仅在当前进程可用。
