# 模型输入

模型输入描述单次模型请求所需的上下文、工具和生成限制。以下为 Agent 内部接口设计，由提供商适配器转换为具体 API 请求。

## 1. 数据结构

```ts
type ToolDefinition = {
  name: string;
  description: string;
  parameters: Record<string, unknown>; // JSON Schema
};

type GenerateInput = {
  model: string;
  messages: Message[];
  tools?: ToolDefinition[];
  maxOutputTokens?: number;
  signal?: AbortSignal;
};
```

`Message` 表示内部统一的消息类型，包含系统指令、用户输入、助手输出和工具结果等语义。它不是直接绑定某家 API 的原始对象。

## 2. 字段含义

| 字段 | 含义 |
|---|---|
| `model` | 本次调用的模型标识 |
| `messages` | 本轮传给模型的消息历史与指令 |
| `tools` | 允许模型请求调用的工具定义，不包含执行函数 |
| `maxOutputTokens` | 单次生成的 token 上限，设置时应为正整数 |
| `signal` | 取消信号，也可由调用方设置超时后触发 |

工具执行函数由 Agent 的工具注册表管理。系统指令和工具结果在不同 API 中可能不是独立消息角色，适配器应按语义转换。

## 3. 请求限制与任务预算

`maxOutputTokens` 只限制单次输出，不限制输入，也不限制整个 Agent 任务的消耗。部分推理模型将推理 token 计入输出预算，适配时需要确认计费与限制口径。

输入还应满足目标模型的上下文限制，并为输出预留空间。Agent 负责选择、裁剪或压缩上下文；不能仅依赖输出上限控制输入大小。

最大循环次数、累计 token 或费用预算和总运行时间属于 Agent 任务限制。每轮请求前检查剩余预算，并据此设置本次输出上限和取消信号。费用估算与实际 token 统计可能存在偏差，预算控制应考虑预留。

## 4. 流式与用量

内部统一返回事件流时，输入无需强制暴露 `stream`。适配器可以使用流式请求，也可将非流式结果转换成相同事件。

Agent 通常需要统计用量，因此适配器应默认请求并收集提供商支持的 usage 信息，再通过输出中的结束事件返回；无需让调用方每轮手动打开统计开关。
