# 流式输入输出

在启动命令前建立所需的管道，并用 `Start()` 启动命令。

- `StdinPipe()` 返回写入端。写完后关闭它，让子进程读到 EOF。
- `StdoutPipe()` 和 `StderrPipe()` 分别返回标准输出和标准错误的读取端。两路都使用管道时，需要并发读取，避免子进程因某一路管道写满而阻塞。

先读完输出管道，再调用 `Wait()`。`Wait()` 会关闭管道，提前调用可能使尚未读完的输出丢失。使用输出管道时应分开调用 `Start()` 和 `Wait()`。

下面的示例在命令运行时将标准输出转发到当前终端。

```go
cmd := exec.Command("git", "status", "--short")
cmd.Stderr = os.Stderr
stdout, err := cmd.StdoutPipe()
if err != nil {
	return err
}
if err := cmd.Start(); err != nil {
	return err
}

_, readErr := io.Copy(os.Stdout, stdout)
if readErr != nil {
	stdout.Close()
}
return errors.Join(readErr, cmd.Wait())
```
