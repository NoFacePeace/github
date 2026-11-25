# os/exec

Go 标准库 `os/exec` 用于启动外部程序、向进程传递参数、读写输入输出并获取退出结果。

## 1. 构造命令

- [构造命令](Construction.md)介绍参数传递、可执行文件查找及工作目录。

## 2. 执行与输入输出

- [执行与输入输出](Execution.md)介绍如何运行命令、收集输出及等待进程退出。

## 3. 超时与错误处理

用 `context.WithTimeout` 派生带超时的上下文，再传给 `exec.CommandContext`。上下文结束时，`CommandContext` 默认终止它直接启动的进程。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

cmd := exec.CommandContext(ctx, "git", "status", "--short")
output, err := cmd.CombinedOutput()
```

- 查找或启动失败时，检查返回的错误，例如程序不在 `PATH` 中或文件没有执行权限。
- 进程启动后以非零状态退出或被信号终止时，返回 `*exec.ExitError`。用 `errors.As` 判断类型，并用 `ExitCode()` 读取退出码（被信号终止时返回 `-1`）。
- 判断上下文是否超时，可用 `errors.Is(ctx.Err(), context.DeadlineExceeded)`。命令被终止时，执行方法本身可能返回 `*exec.ExitError`。

## 4. 参考

- [Go 标准库 `os/exec` 文档](https://pkg.go.dev/os/exec)
