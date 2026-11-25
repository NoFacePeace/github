# Cmd.Start 与 Cmd.Wait

`cmd.Start()` 启动命令，随即返回启动结果。启动成功后，调用方可以继续处理其他工作，之后必须调用一次 `cmd.Wait()`。`Wait()` 等待命令退出、释放相关资源并返回执行结果。

```go
cmd := exec.Command("git", "status", "--short")
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
if err := cmd.Start(); err != nil {
	return err
}

// 可以在这里处理与命令结果无关的工作。
return cmd.Wait()
```
