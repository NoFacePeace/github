# exec.Command

`exec.Command(name, args...)` 把程序名与每个参数分别传入，返回待执行的 `*exec.Cmd`。

```go
cmd := exec.Command("git", "status", "--short")
```

## 1. 查找可执行文件

程序名不含路径分隔符时，`Command` 会用 `exec.LookPath` 查找可执行文件。传入 `./tool` 或绝对路径时，`Command` 直接使用该路径。`PATH` 查找发生在构造时，随后设置 `cmd.Dir` 或 `cmd.Env` 不会重新查找程序。

## 2. 传递参数

`Command` 直接执行程序，`|`、`>` 和 `*` 作为普通参数传递。例如 `exec.Command("find", ".", "-name", "*.go")` 中的 `*.go` 由 `find` 自己匹配。
