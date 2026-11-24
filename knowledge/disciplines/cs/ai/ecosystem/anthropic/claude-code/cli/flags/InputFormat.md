# Claude Code `--input-format`

## 1. 作用

指定 `--print` 模式下 stdin 的输入格式。默认是 `text`，`stream-json` 则按行读取 JSON 消息。

## 2. 用法

使用 `--input-format=stream-json` 时，父进程可以把首轮提示词和追加提示词分别封装为 `user` 消息，写到同一个进程的 stdin。每条消息以换行结束。
