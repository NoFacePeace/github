# Messages API

Messages API 是 Anthropic 为 Claude 模型提供的消息接口，典型端点为 `POST https://api.anthropic.com/v1/messages`。它与 OpenAI Chat Completions 的用途相近，但请求、工具调用及流式事件格式不同，不能直接互换。

## 1. 基本请求

直接调用 Anthropic API 时，使用 `x-api-key` 传递密钥、`anthropic-version` 指定 API 版本。`model` 替换为账号可用的模型 ID；API 版本与模型版本是不同概念。

```http
POST /v1/messages
Host: api.anthropic.com
x-api-key: YOUR_API_KEY
anthropic-version: 2023-06-01
Content-Type: application/json
```

```json
{
  "model": "YOUR_MODEL_ID",
  "max_tokens": 1024,
  "system": "你是一个中文助手。",
  "messages": [
    {"role": "user", "content": "你好"}
  ],
  "stream": false
}
```

`system` 位于顶层，而不是放在 `messages` 中作为 system 消息。响应的 `content` 是内容块数组，可以包含文本和工具调用等不同类型；不能假定整个响应只有一段文本。

## 2. 客户端工具调用

### 2.1 声明工具

在请求顶层增加 `tools`。参数结构使用 `input_schema`，而不是 Chat Completions 的 `function.parameters`。

```json
{
  "tools": [
    {
      "name": "get_weather",
      "description": "查询指定城市的当前天气",
      "input_schema": {
        "type": "object",
        "properties": {
          "city": {"type": "string"}
        },
        "required": ["city"]
      }
    }
  ]
}
```

### 2.2 接收调用并回传结果

模型请求调用工具时，响应 `content` 中包含 `tool_use` 块，`stop_reason` 通常为 `tool_use`。下面仅展示一个内容块：

```json
{
  "type": "tool_use",
  "id": "toolu_001",
  "name": "get_weather",
  "input": {"city": "上海"}
}
```

程序执行工具后，将完整的 assistant 内容加入历史，再通过 user 消息中的 `tool_result` 块回传结果。以下天气为示例数据：

```json
[
  {
    "role": "assistant",
    "content": [
      {
        "type": "tool_use",
        "id": "toolu_001",
        "name": "get_weather",
        "input": {"city": "上海"}
      }
    ]
  },
  {
    "role": "user",
    "content": [
      {
        "type": "tool_result",
        "tool_use_id": "toolu_001",
        "content": "上海晴，25°C"
      }
    ]
  }
]
```

将上面的消息追加到原有 `messages` 后再次请求，保留所需的模型参数与工具定义。结果通过 `tool_use_id` 对应调用的 `id`，不是通过排列顺序匹配。一次出现多个调用时，应为每个调用提供对应结果；工具失败可以在结果块中设置 `is_error: true`。

## 3. 流式输出

设置 `stream: true` 后，响应通过 SSE 分段返回，主要事件包括 `message_start`、`content_block_start`、`content_block_delta`、`content_block_stop`、`message_delta` 与 `message_stop`。

- 文本增量通过 `text_delta` 传递，可直接用于界面展示。
- 工具参数通过 `input_json_delta` 的 `partial_json` 分片传递，需要按内容块 `index` 聚合，收齐后再解析。
- 工具执行由程序管理；模型输出流式化不意味着工具执行结果自动流式化。

适配层应处理流中的错误和异常中断，不能将未收齐的参数当成完整调用执行。

## 4. 与 Chat Completions 的主要差异

| 项目 | Messages | Chat Completions |
|---|---|---|
| 系统提示词 | 顶层 `system` | `messages` 中的 system 消息 |
| 工具参数定义 | `tools[].input_schema` | `tools[].function.parameters` |
| 工具调用 | `content[]` 中的 `tool_use` | assistant 消息中的 `tool_calls[]` |
| 完整调用参数 | `input` 对象 | `function.arguments` JSON 字符串 |
| 工具结果 | user 消息中的 `tool_result` | `role: "tool"` 消息 |
| 结果关联 | `tool_use_id` | `tool_call_id` |
| 流式输出 | 按消息和内容块划分的事件 | `choices[].delta` 增量 |

## 5. 参考资料

- [Messages API](https://platform.claude.com/docs/en/api/messages)
- [工具使用](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview)
- [流式 Messages](https://platform.claude.com/docs/en/build-with-claude/streaming)
