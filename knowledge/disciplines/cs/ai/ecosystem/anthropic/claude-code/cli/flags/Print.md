# Claude Code `-p / --print`

## 1. 作用

以非交互模式运行 Claude Code。可以用 `-p "提示词"` 直接传入输入，也可以让 Claude Code 从 stdin 读取输入。

## 2. 用法

`claude -p "解释当前项目的目录结构"` 适合在脚本中提交一轮任务。程序的输出格式可通过 `--output-format` 选择。
