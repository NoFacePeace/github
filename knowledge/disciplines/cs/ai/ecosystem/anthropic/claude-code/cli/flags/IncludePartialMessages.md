# Claude Code `--include-partial-messages`

## 1. 作用

在 `--print` 和 `--output-format=stream-json` 模式下加入生成中的消息片段。

## 2. 用法

父进程可以在完整回复产生之前收到增量，逐步更新界面。最终执行结果仍从 `result` 事件读取。
