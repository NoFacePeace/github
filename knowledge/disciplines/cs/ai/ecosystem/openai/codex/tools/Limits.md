# Codex 工具输出限制

本文区分官方指南中的推荐策略与产品配置中明确提供的能力。核对日期为 2026-10-20；不能把指南中的数值视为所有客户端和工具固定使用的默认值。

## 1. 检查时机

限制针对单次工具调用返回的结果，在结果进入后续模型上下文或历史时应用，不需要等整个上下文接近上限。具体工具还可能提供自己的行数、分页或输出长度参数。

## 2. 官方推荐的截断策略

OpenAI 的 Codex Prompting Guide 推荐：

- 将单次工具输出限制在约 10,000 Token。
- 可以用字节数除以 4 粗略估算 Token 数；这不是精确分词，非 ASCII 内容可能有较大误差。
- 超限后，展示预算一半用于开头，一半用于结尾。
- 在中间加入标记，说明省略的 Token 数量。

```text
[开头内容]
…N tokens truncated…
[结尾内容]
```

这是面向构建者的推荐方式，不能单凭这段文档确定某个 Codex 版本、桌面工具或连接器的实际阈值。

## 3. 配置入口

官方配置参考提供 `tool_output_token_limit`，用于设置单条工具或函数输出存入历史的 Token 预算。例如：

```toml
tool_output_token_limit = 10000
```

示例值不代表所有模型和版本的默认值。具体工具的输出参数、客户端的返回限制与历史预算也不一定是同一层限制。

## 4. 截断与外置

截断只规定模型保留哪些片段，不自动保证全文保存到可读取文件。官方上述截断建议本身没有提供统一的“保存全文并返回路径”保证。

如果任务需要恢复完整输出，可以在执行命令时主动将全文保存为文件，再读取必要片段。已经进入历史的全文不会仅因另存为文件就自动减少占用，仍需运行时调整后续模型输入。

## 5. 参考资料

- [Tool Response Truncation](https://developers.openai.com/cookbook/examples/gpt-5/codex_prompting_guide#tool-response-truncation)
- [Codex 配置参考](https://learn.chatgpt.com/docs/config-file/config-reference)
