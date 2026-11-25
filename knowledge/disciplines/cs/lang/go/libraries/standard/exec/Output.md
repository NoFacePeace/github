# Cmd.Output 与 Cmd.CombinedOutput

两个方法都会启动命令并等待它退出，以 `[]byte` 返回收集到的内容。

- `cmd.Output()` 返回标准输出，调用前需保持 `cmd.Stdout` 为 `nil`。要分别获取两路输出，可把 `cmd.Stderr` 设为 `*bytes.Buffer`，再从 `Output()` 的返回值取标准输出，从缓冲区取标准错误。若命令以非零状态退出且 `cmd.Stderr` 未设置，可从 `*exec.ExitError.Stderr` 读取捕获的标准错误。
- `cmd.CombinedOutput()` 将标准输出和标准错误收集到同一个缓冲区。调用前需将 `cmd.Stdout`、`cmd.Stderr` 保持为 `nil`。

下面的示例在当前 Git 仓库运行 `git status --short`。命令失败时，将已收集的输出和错误一起返回。

```go
cmd := exec.Command("git", "status", "--short")
output, err := cmd.CombinedOutput()
if err != nil {
	return fmt.Errorf("git status 执行失败：%w\n%s", err, output)
}
fmt.Print(string(output))
return nil
```
