# panic

`panic` 是 Go 的预声明内建函数，不需要导入包，也不是关键字。调用 `panic(v)` 会中断当前正常执行流程，开始执行当前 goroutine 调用栈中已登记的 defer；如果最终没有被恢复，整个 Go 程序会退出。

## 1. 基本用法

panic 接收一个值，该值可以是字符串、error 或其他类型，用来描述异常情况。panic 没有正常返回值，调用后的普通语句不会继续执行。

```go
package main

import "fmt"

func work() {
    defer fmt.Println("work 清理")
    panic("unexpected state")
}

func main() {
    defer fmt.Println("main 清理")
    work()
    fmt.Println("不会执行")
}
```

程序依次打印 `work 清理`、`main 清理`，随后运行时报告未恢复的 panic 并退出。示例中的 defer 使用标准输出，panic 诊断通常写入标准错误；两种输出合并显示时，显示顺序可能受缓冲和采集方式影响。

panic 值不一定实现 error 接口。需要保留错误信息时，可以传入 error 值，例如 `panic(fmt.Errorf("invalid state: %s", state))`，但普通业务失败通常更适合直接返回 error。

## 2. 传播与 defer 执行顺序

1. 当前函数停止正常执行，按登记顺序的逆序执行其 defer。
2. 如果没有恢复，继续向当前 goroutine 的调用方展开，并执行相应函数的 defer。
3. 如果某个符合条件的 deferred 函数成功 recover，停止此次 panic 的传播；登记该 defer 的函数执行剩余 defer 后返回。
4. 如果传播到当前 goroutine 的调用栈顶仍未恢复，运行时报告 panic 并结束整个程序。

恢复不会让程序回到 panic 发生的位置，也不会自动回滚已经完成的状态修改。defer 中还可能发生新的 panic，因此清理代码也应保持可靠。

## 3. 主动调用与运行时 panic

panic 既可以由代码主动触发，也可以由运行时检测到某些非法操作时触发：

| 来源 | 示例 |
| --- | --- |
| 主动触发 | `panic("invalid internal state")` |
| 下标越界 | 对长度为 1 的切片访问 `values[1]` |
| nil 指针解引用 | 对 nil 的 `*int` 执行 `*p` |
| nil map 写入 | 对未初始化的 map 执行 `m["key"] = 1` |
| 类型断言失败 | 使用不带 ok 的 `value.(T)`，且动态类型不匹配 |
| channel 误用 | 向已关闭的 channel 发送、重复关闭 channel 或关闭 nil channel |

这些示例指运行时发生的非法操作。编译期能够确定的部分错误会直接被编译器拒绝，例如数组的常量下标越界。

并非所有运行时错误都是可恢复的 panic。例如并发 map 误用、对未锁定的 Mutex 执行 Unlock，可能触发运行时致命错误，不能依靠 recover 处理。

## 4. 与 goroutine 和 recover 的关系

panic 只沿发生异常的 goroutine 的调用栈传播，不会沿启动关系传给父 goroutine。主 goroutine 中的 recover 无法捕获子 goroutine 的 panic。

但影响范围不限于当前 goroutine：任意 goroutine 的 panic 如果最终未恢复，都会导致整个进程退出。其他 goroutine 的 defer 不会因此被统一执行。

需要隔离请求或任务失败时，应在相应 goroutine 的任务边界设置恢复逻辑。具体调用条件、日志示例和任务内外的放置建议参见 [recover](Recover.md)。

## 5. 什么时候使用

| 情况 | 通常的选择 |
| --- | --- |
| 文件不存在、网络超时、用户参数错误 | 返回 error，由调用方决定如何处理 |
| 内部不变量被破坏，当前执行路径无法合理继续 | 可以考虑 panic，并排查根因 |
| Must 风格的初始化辅助函数 | 可在失败时 panic，但应通过命名和文档明确约定 |
| 隔离请求或独立任务中的意外 panic | 在边界使用 recover，记录堆栈并标记失败 |

例如标准库的 `regexp.MustCompile` 会在表达式无效时 panic，适合初始化代码中由开发者提供的固定表达式。处理用户输入的表达式时，通常应使用返回 error 的 `regexp.Compile`。

不应把 panic 当作日常的分支、循环退出或业务错误传递方式。使用 recover 后能否继续运行，取决于共享数据和业务状态是否仍然可信。

一种有明确边界的用法是：包内部在深层调用中用特定私有类型的 panic 展开调用栈，对外入口只把该类型转成 error，其他意外 panic 重新抛出。标准库 encoding/json 的编码实现采用了这一模式；它不意味着应把所有业务函数改成 panic/recover 风格。

## 6. 与其他退出方式的区别

| 方式 | 行为 | 是否执行 defer | 能否被 recover 恢复 |
| --- | --- | --- | --- |
| `return` | 正常返回当前函数；main 返回时程序结束 | 执行当前函数的 defer | 不适用 |
| `panic(v)` | 展开当前 goroutine 的调用栈，未恢复则结束程序 | 执行展开路径上已登记的 defer | 满足条件时可以 |
| `os.Exit(code)` | 直接结束整个进程 | 不执行 | 不能 |
| 标准库 `log.Fatal` | 输出日志，然后调用 `os.Exit(1)` | 不执行 | 不能 |

main 返回不会等待其他 goroutine 完成，也不会自动执行它们的 defer。

Go 1.21 起，在默认行为下，`panic(nil)` 会触发非 nil 的 `*runtime.PanicNilError`。它不表示“不发生 panic”；旧版本及 `GODEBUG=panicnil=1` 的兼容行为存在差异。

## 7. 实现与源码入口

- `src/builtin/builtin.go`：panic 的文档化声明，不是普通函数实现。
- `src/runtime/panic.go`：panic 展开、defer 执行和恢复的主要运行时逻辑。
- `src/cmd/compile/internal`：panic 调用及相关运行时检查的编译处理。
