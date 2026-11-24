# Claude Code CLI

## 1. 单次执行

默认运行 `claude` 会进入交互会话。脚本或父进程可以通过 `-p`（`--print`）在非交互模式下提交一轮提示词，并从 stdout 读取输出。

```sh
claude -p "解释当前项目的目录结构" \
  --output-format=stream-json \
  --verbose \
  --include-partial-messages
```

`stream-json` 输出采用 JSONL。每行是一条独立的 JSON 事件，调用方可以边运行边读取。开启部分消息后，输出流还包含生成中的回复片段。

## 2. 流式输入

在非交互模式下加上 `--input-format=stream-json`，Claude Code 会从 stdin 逐行读取 JSON 消息。

```sh
printf '%s\n' '{"type":"user","message":{"role":"user","content":"解释当前项目的目录结构"}}' |
  claude -p --input-format=stream-json --output-format=stream-json \
    --verbose --include-partial-messages
```

示例中的 `printf` 写完一条消息后关闭管道。父进程也可以保持写端开启，把首轮提示词和后续追加内容作为独立的 `user` 消息写给同一个进程。这种在执行中追加输入的用法也称为 steer。

## 3. 参数

[参数总表](flags/Flags.md) 汇总常用参数及各自的用法。

## 4. 会话续接

使用 `--resume <会话 ID>` 指定要续接的 Claude Code 会话。本轮提示词可以继续通过 `-p` 参数或流式 stdin 输入传入。

```sh
claude -p --resume "<会话 ID>" "结合上一轮内容继续解释"
```

## 5. 官方文档

- [CLI 参考](https://code.claude.com/docs/en/cli-reference)
