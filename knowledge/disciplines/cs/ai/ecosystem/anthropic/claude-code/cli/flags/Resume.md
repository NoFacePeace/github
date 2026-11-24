# Claude Code `--resume`

## 1. 作用

按会话 ID 续接已有的 Claude Code 会话。短写为 `-r`。

## 2. 用法

例如 `claude -p --resume "<会话 ID>" "结合上一轮内容继续解释"`。在流式输入模式下，也可以用同一个 `--resume` 参数选择会话，再通过 stdin 发送本轮消息。
