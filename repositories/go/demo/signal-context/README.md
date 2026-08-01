# 使用 signal.NotifyContext 优雅退出

在 `repositories/go` 目录下运行示例：

```sh
go run ./demo/signal-context
```

按下 `Ctrl-C`（或发送 `SIGTERM`）后，`context` 会被取消，`run` 随即优雅
退出。`run` 是负责等待 `ctx.Done()` 的阻塞函数，因此 `main` 不需要再等待
该 `context`。

`defer stop()` 会在 `main` 返回时取消信号注册。这样信号处理器的生命周期由
调用方控制，而不是隐藏在包级全局状态中。

示例中的 goroutine 会在第一次信号取消 `context` 后立刻调用 `stop()`，恢复
操作系统的默认信号行为。因此服务仍在清理时再次按下 `Ctrl-C`，进程会被直接
终止；无需在业务代码中调用 `os.Exit`。
