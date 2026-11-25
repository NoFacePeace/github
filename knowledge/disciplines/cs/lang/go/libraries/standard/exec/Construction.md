# 构造命令

构造命令得到待执行的 `*exec.Cmd`。启动前可以设置 `cmd.Dir` 指定子进程的工作目录，设置 `cmd.Env` 指定环境变量。`cmd.Env` 为 `nil` 时，子进程继承当前进程的环境。

## 1. 构造方式

- [Command](Command.md)按程序名和参数构造命令。
- [CommandContext](CommandContext.md)将命令与上下文的取消和截止时间关联。
- [手动构造 Cmd](Cmd.md)直接设置可执行文件路径和参数列表。
