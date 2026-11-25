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

## 1. 用 Stdout 收集输出

在 `Start()` 前将 `cmd.Stdout` 设为 `*bytes.Buffer`，Go 会在命令运行期间把标准输出的字节写入缓冲区。`Start()` 返回时，缓冲区可能只有部分内容。`bytes.Buffer` 不支持并发读写，应在 `Wait()` 返回后再用 `String()` 或 `Bytes()` 读取。

## 2. 用 Stdout 按行输出

需要在运行期间收到换行符就输出时，可以把 `cmd.Stdout` 设为自定义的 `io.Writer`。一次 `Write` 可能收到半行或多行，写入器要保留未换行的尾部内容。下面的写入器只打印以 `\n` 结束的行。

```go
type lineWriter struct {
	pending bytes.Buffer
}

func (w *lineWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		line, rest, found := bytes.Cut(p, []byte{'\n'})
		w.pending.Write(line)
		if !found {
			return written + len(line), nil
		}
		if _, err := fmt.Fprintln(os.Stdout, w.pending.String()); err != nil {
			return written, err
		}
		w.pending.Reset()
		written += len(line) + 1
		p = rest
	}
	return written, nil
}
```

在启动前设置 `cmd.Stdout = &lineWriter{}`。如果还要打印末尾没有换行的内容，可以保留写入器的引用，并在 `Wait()` 返回后输出 `pending` 中剩余的内容。
