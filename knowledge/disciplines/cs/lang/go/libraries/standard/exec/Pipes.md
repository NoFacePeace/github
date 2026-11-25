# 流式输入输出

在启动命令前建立所需的管道，并用 `Start()` 启动命令。

- `StdinPipe()` 返回写入端。写完后关闭它，让子进程读到 EOF。
- `StdoutPipe()` 和 `StderrPipe()` 分别返回标准输出和标准错误的读取端。两路都使用管道时，需要并发读取，避免子进程因某一路管道写满而阻塞。

## 1. 读取输出

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

## 2. 按行读取输出

管道按字节读取，一次 `Read` 可能只读到半行或半个 UTF-8 字符。要按行处理，可以把上例中 `io.Copy` 及其后的代码换成 `bufio.Scanner`。

```go
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
	fmt.Println(scanner.Text())
}

scanErr := scanner.Err()
if scanErr != nil {
	stdout.Close()
}
return errors.Join(scanErr, cmd.Wait())
```

`scanner.Text()` 去掉行尾换行符。末尾没有换行的内容也会在读到 EOF 时作为一行返回。实时显示取决于子进程何时把数据写入管道。如果输出可能有超长行，在开始扫描前用 `scanner.Buffer` 调整最大长度。

## 3. 启动后写入输入

在 `Start()` 前调用 `StdinPipe()`，启动后用 `Write` 或 `fmt.Fprintln` 发送内容。可以连续写入多次，发完后关闭管道，让子进程读到 EOF。

```go
cmd := exec.Command("cat")
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
stdin, err := cmd.StdinPipe()
if err != nil {
	return err
}
if err := cmd.Start(); err != nil {
	return err
}

_, writeErr := fmt.Fprintln(stdin, "你好")
closeErr := stdin.Close()
return errors.Join(writeErr, closeErr, cmd.Wait())
```

如果要让用户在终端键入内容，可以在启动前设置 `cmd.Stdin = os.Stdin`。如果程序需要一边发送输入一边读取输出，应并发处理管道读写，避免互相等待。
