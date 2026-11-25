# 手动构造 exec.Cmd

可以直接用结构体字面量构造 `*exec.Cmd`。`Path` 指定要执行的文件，`Args` 保存完整的参数列表，通常以程序名作为 `Args[0]`。

```go
path, err := exec.LookPath("git")
if err != nil {
	return err
}

cmd := &exec.Cmd{
	Path: path,
	Args: []string{"git", "status", "--short"},
}
return cmd.Run()
```

直接构造 `Cmd` 时，可以像示例一样用 `exec.LookPath` 查找程序，也可以把已知的可执行文件路径直接填入 `Path`。如果 `Args` 为空，执行时会用 `Path` 作为 `Args[0]`。
