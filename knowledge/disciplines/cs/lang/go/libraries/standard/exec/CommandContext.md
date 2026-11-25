# exec.CommandContext

`exec.CommandContext(ctx, name, args...)` 构造与 `context.Context` 关联的命令。程序名、参数和路径查找规则与 `exec.Command` 相同。

## 1. 设置超时

用 `context.WithTimeout` 派生带超时的上下文。下面的代码为 `git status --short` 设置五秒超时。

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

cmd := exec.CommandContext(ctx, "git", "status", "--short")
output, err := cmd.CombinedOutput()
```

## 2. 取消行为

命令运行期间上下文结束时，`CommandContext` 默认调用 `Process.Kill` 终止它直接启动的进程。
