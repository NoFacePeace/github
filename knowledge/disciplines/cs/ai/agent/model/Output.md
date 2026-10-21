# 模型输出

模型输出统一为事件流，供 Agent 收集助手消息和工具调用，同时供界面展示文本增量。这是内部接口约定，不是提供商的原生事件格式。

## 1. 数据结构

```ts
type ToolCall = {
  id: string;
  name: string;
  arguments: string;
};

type Usage = {
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
  cacheReadTokens?: number;
  cacheWriteTokens?: number;
};

type ModelEvent =
  | { type: "text_delta"; text: string }
  | { type: "tool_call_ready"; call: ToolCall }
  | {
      type: "turn_finished";
      reason: "stop" | "tool_calls" | "length" | "other";
      usage?: Usage;
    };

interface ModelProvider {
  generate(input: GenerateInput): AsyncIterable<ModelEvent>;
}
```

`GenerateInput` 是单次请求的输入类型。`arguments` 为完整 JSON 字符串，执行工具前还需要解析和校验。

## 2. 事件含义

| 事件 | 处理方式 |
|---|---|
| `text_delta` | 立即展示并按顺序拼接文本 |
| `tool_call_ready` | 收集一个完整工具调用，通过 ID 关联后续结果 |
| `turn_finished` | 检查结束原因、记录用量并决定是否继续 |

适配器内部聚合工具参数分片，不将尚未完整的参数交给 Agent 执行。一个响应可以产生多个工具调用事件；工具调用生成完成不表示工具执行完成。

## 3. 用量统计

usage 表示单次模型调用的实际用量，不是请求预算。字段可选，提供商未返回的值应保留为缺失，不能当作零。

缓存统计口径各家不同：缓存 token 可能已经包含在输入 token 中，也可能单独列出。适配器应明确转换口径，不能简单把缓存数量重复加到输入或总量上。费用应按提供商定价与统计口径另行计算。

流式 usage 可能在生成结束标记之后的独立片段中到达。适配器应收齐预期的结束信息与用量后，再发出 `turn_finished`。Agent 按每次调用累计用量，避免将提供商的最终统计与中间累计值重复相加；统计不完整时应保留缺失状态。

## 4. 完成与错误

每次正常结束的调用发出一个 `turn_finished`；`length` 表示生成被长度限制截断，`other` 需要单独判断，均不能直接视为可执行工具的成功状态。

网络、协议和服务端错误通过异步迭代器抛出异常，不伪造成正常结束。消费者应完整消费迭代器，确认没有异常后再执行收集到的调用。异常中断时可能只有部分文本，usage 也可能无法获得。
