# Claude Code CLI 参数

## 1. 常用参数

| 参数 | 作用 |
| --- | --- |
| [`-p`、`--print`](Print.md) | 启用非交互模式，可通过参数或 stdin 提供提示词。 |
| [`--input-format=stream-json`](InputFormat.md) | 从 stdin 读取 JSONL 消息。 |
| [`--output-format=stream-json`](OutputFormat.md) | 将运行事件以 JSONL 写入 stdout。 |
| [`--verbose`](Verbose.md) | 开启详细输出，配合 `--print` 的流式 JSON 输出使用。 |
| [`--include-partial-messages`](IncludePartialMessages.md) | 在流式输出中包含生成中的回复片段。 |
| [`--model <模型>`](Model.md) | 指定本轮使用的模型。 |
| [`--effort <强度>`](Effort.md) | 设置本轮推理强度。 |
| [`--debug`](Debug.md) | 开启调试日志。 |
| [`--append-system-prompt-file <文件>`](AppendSystemPromptFile.md) | 将文件内容追加到默认系统提示词。 |
| [`--disallowedTools <工具列表>`](DisallowedTools.md) | 禁用列出的工具，支持逗号分隔。 |
| [`--dangerously-skip-permissions`](DangerouslySkipPermissions.md) | 跳过 Claude Code 的权限检查。 |
| [`--resume <会话 ID>`](Resume.md) | 指定要续接的 Claude Code 会话。 |
