# Claude Code `--output-format`

## 1. 作用

指定 `--print` 模式的输出格式。`text` 输出文本，`json` 输出结构化结果，`stream-json` 在执行期间逐行输出 JSON 事件。

## 2. 用法

需要边执行边读取时，可以用 `claude -p "解释当前项目" --output-format=stream-json --verbose`。JSONL 的一行就是一个事件。调用方可以根据 `type` 字段分别处理初始化、消息和结果。
