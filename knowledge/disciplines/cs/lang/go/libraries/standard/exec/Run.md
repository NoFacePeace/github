# Cmd.Run

`cmd.Run()` 启动命令并等待它退出，返回启动或执行过程中发生的错误。可以在启动前设置 `cmd.Stdin`、`cmd.Stdout` 和 `cmd.Stderr`。

下面的示例将命令的标准输入、标准输出和标准错误接到当前终端。

```go
cmd := exec.Command("git", "status", "--short")
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
return cmd.Run()
```
