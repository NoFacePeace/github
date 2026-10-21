# turn_finished

`turn_finished` 表示本轮模型生成结束，并提供统一的结束原因。一轮模型生成结束不等于整个 Agent 任务完成，后续可能还需要执行工具并再次请求模型。

## 1. 数据结构

```ts
type Usage = {
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
  cacheReadTokens?: number;
  cacheWriteTokens?: number;
};

type TurnFinished = {
  type: "turn_finished";
  reason: "stop" | "tool_calls" | "length" | "other";
  usage?: Usage;
};
```

## 2. 结束原因

| reason | 含义 |
|---|---|
| `stop` | 模型正常停止生成 |
| `tool_calls` | 模型生成工具调用，需要程序处理结果并继续 |
| `length` | 达到输出长度限制，结果可能不完整 |
| `other` | 未归入上述类型的其他结束原因，需要单独处理 |

结束原因由适配器根据原始响应映射。不要把未知原因默认为正常完成，也不要把 `length` 当作工具参数已经完整的依据。

## 3. 事件流约定

一次正常结束的模型调用发出一个 `turn_finished`，之后不再产生本轮的文本或工具调用事件。消费者仍应将异步迭代器完整消费结束，以捕获后续传输或解析异常。

网络、协议或服务端错误通过迭代器抛出异常，不应伪造一个正常结束事件。Agent 结合结束原因、收集到的文本与工具调用决定下一步动作。

## 4. 用量统计

`usage` 记录单次调用的实际用量。提供商未返回的字段应保持缺失，不能当作零；缓存统计是否包含在输入 token 中应由适配器按明确口径处理。

流式用量可能在单独的尾部片段中返回。适配器收齐结束信息与用量后再发出本事件，Agent 按调用累计，避免重复统计。
