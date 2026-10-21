# Responses API

Responses 使用不同类型的 Item 表示模型输入与输出，端点为 `POST /v1/responses`。消息、函数调用和函数结果是不同类型的 Item，而不是全部放进一条 assistant 消息。

## 1. 基本请求与响应

直接调用 OpenAI 时，地址为 `https://api.openai.com/v1/responses`，请求头包含 `Authorization: Bearer YOUR_API_KEY` 与 `Content-Type: application/json`。模型 ID 需要替换为可用模型。

```json
{
  "model": "YOUR_MODEL_ID",
  "instructions": "你是一个中文助手。",
  "input": "你好",
  "stream": false
}
```

响应通过 `output[]` 返回带类型的 Item，不能假定每个 Item 都是文本消息。官方 SDK 的 `response.output_text` 可提取文本；它是 SDK 辅助属性，不应假设原始 HTTP JSON 一定包含同名字段。

## 2. 客户端工具调用

### 2.1 定义工具

在请求中增加 `tools`。函数名称和参数定义直接位于工具对象上，没有 Chat Completions 中的 `function` 包装层：

```json
{
  "model": "YOUR_MODEL_ID",
  "input": "上海天气怎么样？",
  "tools": [
    {
      "type": "function",
      "name": "get_weather",
      "description": "查询指定城市的当前天气",
      "parameters": {
        "type": "object",
        "properties": {"city": {"type": "string"}},
        "required": ["city"],
        "additionalProperties": false
      },
      "strict": true
    }
  ],
  "tool_choice": "auto"
}
```

### 2.2 接收调用并回传结果

响应的 `output[]` 中可能包含如下 Item：

```json
{
  "type": "function_call",
  "id": "fc_weather_001",
  "call_id": "call_weather_001",
  "name": "get_weather",
  "arguments": "{\"city\":\"上海\"}"
}
```

程序执行工具后，在下一轮 `input` 中提交结果 Item。天气为示例数据：

```json
{
  "type": "function_call_output",
  "call_id": "call_weather_001",
  "output": "{\"city\":\"上海\",\"temperature_c\":25,\"condition\":\"晴\"}"
}
```

结果匹配调用的 `call_id`，不是 Item 的 `id`。可以通过 `previous_response_id` 引用上一轮响应并提交结果，也可以自行将上一轮完整 `output` 加入历史后追加结果；手动管理历史时不要遗漏需要保留的推理项或工具项。

## 3. 对话上下文

| 方式 | 使用方法 |
|---|---|
| 手动管理 | 将历史输入与返回的相关 Item 传入下一轮 `input` |
| 响应链 | 通过 `previous_response_id` 引用上一轮响应 |
| Conversations | 使用持久化会话对象管理上下文 |

`previous_response_id` 不会自动继承上一轮顶层 `instructions`，稳定的指令应按需要重新发送；它也不会免除历史输入的 token 计费。

## 4. 流式输出

设置 `stream: true` 后通过 SSE 接收带类型的事件，常见事件包括：

- `response.output_text.delta`：文本增量。
- `response.output_item.added`：增加输出 Item。
- `response.function_call_arguments.delta`：工具参数增量。
- `response.function_call_arguments.done`：参数生成完成。
- `response.output_item.done`：输出 Item 完成。
- `response.completed`：响应成功完成。

适配层按事件携带的输出索引与 Item 标识聚合工具调用，同时保留结果关联所需的 `call_id`。还应处理失败、不完整响应和网络中断，不能仅收到参数增量就执行工具。

## 5. 与 Chat Completions 的区别

| 项目 | Responses | Chat Completions |
|---|---|---|
| 输入 | `input` | `messages` |
| 输出 | 带类型的 `output[]` | `choices[].message` |
| 函数定义 | `tools[].name`、`parameters` | `tools[].function.name`、`parameters` |
| 函数结果 | `function_call_output` Item | tool 消息 |
| 调用关联 | `call_id` | `tool_call_id` 对应调用 `id` |
| 结构化输出 | `text.format` | `response_format` |
| 流式 | 带类型的事件 | `choices[].delta` |

Chat Completions 仍受支持。迁移时应同时调整请求结构、结果解析、工具循环、上下文管理与流式处理；第三方兼容程度需要单独确认。

## 6. 参考资料

- [Responses API](https://developers.openai.com/api/docs/guides/responses)
- [Function calling](https://developers.openai.com/api/docs/guides/function-calling)
- [迁移到 Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses)
