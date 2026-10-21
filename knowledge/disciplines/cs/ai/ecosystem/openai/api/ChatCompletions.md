# Chat Completions API

Chat Completions 使用消息列表调用模型，端点为 `POST /v1/chat/completions`。第三方所谓“OpenAI 兼容”通常指这种格式，但具体能力需要单独确认。

## 1. 基本请求与响应

直接调用 OpenAI 时，地址为 `https://api.openai.com/v1/chat/completions`，请求头包含 `Authorization: Bearer YOUR_API_KEY` 与 `Content-Type: application/json`。下面的模型 ID 需要替换为支持所需功能的可用模型。

```json
{
  "model": "YOUR_MODEL_ID",
  "messages": [
    {"role": "system", "content": "你是一个中文助手。"},
    {"role": "user", "content": "你好"}
  ],
  "stream": false
}
```

非流式响应通过 `choices[]` 返回生成结果，文本通常位于 `choices[0].message.content`。assistant 消息也可能包含工具调用，不能只读取文本。

## 2. 工具调用

### 2.1 请求工具调用

在请求顶层添加 `tools`。以下请求声明天气工具，`tool_choice: "auto"` 允许模型自行决定是否调用：

```json
{
  "model": "YOUR_MODEL_ID",
  "messages": [
    {"role": "user", "content": "上海天气怎么样？"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "查询指定城市的当前天气",
        "parameters": {
          "type": "object",
          "properties": {"city": {"type": "string"}},
          "required": ["city"],
          "additionalProperties": false
        }
      }
    }
  ],
  "tool_choice": "auto",
  "stream": false
}
```

### 2.2 执行并回传结果

模型请求调用工具时，assistant 消息可能如下，响应的 `finish_reason` 通常为 `tool_calls`：

```json
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_weather_001",
      "type": "function",
      "function": {
        "name": "get_weather",
        "arguments": "{\"city\":\"上海\"}"
      }
    }
  ]
}
```

程序解析 `arguments` JSON 字符串并执行工具，将上面的 assistant 消息加入原有历史，再追加工具结果消息。天气为示例数据：

```json
{
  "role": "tool",
  "tool_call_id": "call_weather_001",
  "content": "{\"city\":\"上海\",\"temperature_c\":25,\"condition\":\"晴\"}"
}
```

携带完整 `messages` 和所需工具定义再次请求模型，模型可以继续调用工具或生成最终回答。每个结果通过 `tool_call_id` 匹配调用 `id`，不是按顺序或函数名匹配；同一轮多个调用应各有对应结果。

模型只生成调用请求，客户端工具由程序执行。不同模型对工具调用与推理参数组合的支持可能不同，应按模型文档确认。

## 3. 流式输出

设置 `stream: true` 后，服务通过 SSE 返回多个 JSON 片段，主要读取 `choices[].delta`，而不是非流式的 `choices[].message`。

- `delta.content`：新增文本，可逐段展示。
- `delta.tool_calls`：工具调用分片，按 choice 的 `index` 和工具调用的 `index` 分别聚合。
- `function.arguments`：每次可能只包含部分 JSON，需要依次拼接；调用 ID、类型和名称可能只在初始片段出现。
- `finish_reason`：标记该 choice 的结束原因；流通常以 `data: [DONE]` 结束。

收齐工具参数并确认本轮正常完成后再执行工具，避免执行截断或异常中断的调用。工具回传后的下一轮模型请求仍可使用流式；工具自身能否输出进度取决于工具实现。

## 4. 在 Agent 中使用

适配层可以将文本增量、完整工具调用和本轮完成转换为统一事件。Agent 核心收集 assistant 消息，执行工具并追加结果，再进入下一轮。非流式响应也可转换成相同事件，复用工具循环。

## 5. 参考资料

- [Chat Completions API Reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create)
- [Function calling](https://developers.openai.com/api/docs/guides/function-calling)
- [Responses API 对比与迁移](https://developers.openai.com/api/docs/guides/migrate-to-responses)
