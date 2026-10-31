# recover

`recover` 是 Go 的预声明内建函数，用于恢复当前 goroutine 中正在发生的 panic。它通常与 defer 配合使用，不是关键字，也不是处理普通业务错误的首选方式。

## 1. 基本行为

```go
package main

import "fmt"

func run() {
    defer func() {
        if v := recover(); v != nil {
            fmt.Println("捕获 panic:", v)
        }
    }()

    panic("发生错误")
    // 这里及其后的普通语句不会执行。
}

func main() {
    run()
    fmt.Println("main 继续执行")
}
```

输出：

```text
捕获 panic: 发生错误
main 继续执行
```

panic 会中断当前正常执行流程，沿当前 goroutine 的调用栈执行已登记的 defer。成功 recover 后，不会回到 panic 发生的位置继续执行；登记该 defer 的函数会执行剩余的 defer，然后返回调用方。

## 2. 调用规则

- 必须由 deferred 函数直接调用 recover，才能恢复当前 panic。
- 只能恢复同一 goroutine 的 panic，无法跨 goroutine 捕获。
- 没有正在发生的 panic，或不满足调用条件时，recover 返回 nil。
- 任意 goroutine 的 panic 如果最终没有被恢复，都会导致整个 Go 程序退出。

下面的辅助函数多了一层普通调用，不能恢复 panic：

```go
func tryRecover() {
    recover()
}

func invalid() {
    defer func() {
        tryRecover() // recover 不是由 deferred 函数直接调用。
    }()
    panic("未被恢复")
}
```

如果改成 `defer tryRecover()`，tryRecover 本身就是 deferred 函数，其中直接调用 recover 可以生效。单独写 `defer recover()` 则不能恢复 panic。

Go 1.21 起，在默认行为下，`panic(nil)` 会产生非 nil 的运行时 panic 值，因此有效的 recover 仍能用 `v != nil` 判断。旧版本或启用 `GODEBUG=panicnil=1` 的兼容行为时，不能依赖这一保证。

## 3. 适用场景

| 场景 | 恢复后的处理 |
| --- | --- |
| HTTP 请求边界 | 记录堆栈；响应尚未写出时返回 500 |
| 独立后台任务、定时任务或消息消费 | 标记本次任务失败，根据业务规则决定重试或继续后续任务 |
| 插件或第三方代码调用边界 | 记录异常，按接口约定转换为 error 或终止本次调用 |

参数错误、文件不存在、数据库超时等预期内的失败，通常应返回和处理 error。recover 不会自动回滚数据、重试任务或恢复被破坏的业务状态。

继续运行的前提是程序状态仍然可信。如果共享数据只更新了一半，或者关键约束已经被破坏，应先处理一致性问题；无法安全恢复时，让进程退出并由外部机制重启可能更合适。

## 4. 放在任务内部还是外部

### 4.1 通常放在外层任务边界

由负责启动和管理任务的代码统一捕获 panic，任务函数专注业务逻辑。下面用 task 参数表示具体任务：

```go
func startTask(task func()) {
    go func() {
        defer func() {
            if v := recover(); v != nil {
                log.Printf("任务 panic: %v\n%s", v, debug.Stack())
                // 在此按业务规则标记任务失败。
            }
        }()

        task()
    }()
}
```

示例需要导入 `log` 和 `runtime/debug`。应在恢复现场记录堆栈，便于定位 panic 来源。其他地方调用同一个任务函数时，可以采用各自的异常处理策略。

对于持续消费或循环执行任务的 worker，应在每次任务调用的函数边界恢复。仅在整个 worker 函数入口登记 defer，发生 panic 后会使该 worker 返回，不会自动继续下一轮循环。

### 4.2 包内部只转换约定的 panic

Effective Go 和 Go 官方博客认可一种用法：包内部在深层调用中使用 panic 展开调用栈，在对外入口恢复并返回 error。这是一种内部实现选择，普通业务函数仍应优先直接返回 error。

建议使用包私有类型标记有意触发的 panic，只转换该类型，其他 panic 用 `panic(v)` 重新抛出。标准库 encoding/json 的编码实现就使用私有的 jsonError 包装类型作此区分。

下面的 doTask 代表包内的具体任务逻辑；它可以直接返回 error，也可以在内部深层调用中触发约定的 taskFailure：

```go
type taskFailure struct {
    err error
}

func DoTask() (err error) {
    defer func() {
        if v := recover(); v != nil {
            if failure, ok := v.(taskFailure); ok {
                err = failure.err
            } else {
                panic(v) // 非约定的 panic 继续向外传播。
            }
        }
    }()

    return doTask()
}

// 包内深层函数必要时调用，err 应为非 nil：
// panic(taskFailure{err: err})
```

使用命名返回值保存错误，避免恢复后返回零值 nil，让调用方误以为成功。如果还有其他结果返回值，应清空无效结果，避免返回部分完成的数据。

仅判断 `v.(error)` 无法识别约定的内部失败：数组越界、nil 指针解引用等运行时 panic 值也可能实现 error 接口。不应把这些 bug 在普通业务函数中无差别转换成正常失败。

### 4.3 任务隔离边界可以捕获所有可恢复的 panic

最外层请求处理器或独立任务执行器的职责可能就是隔离意外失败。此时可以像前面的 startTask 一样捕获所有满足语言恢复条件的 panic，也可以将它们转换为 error 交给任务管理代码，但应满足以下条件：

- 在恢复现场记录或保存 `debug.Stack()`，普通的 `fmt.Errorf` 不会自动保留 panic 堆栈。
- 明确标识这是意外 panic，例如使用专门的错误类型，避免与普通业务失败混淆。
- 标记本次任务失败；是否重试取决于任务幂等性、已产生的副作用及状态一致性，不能只因转成 error 就自动重试。
- 确认继续处理其他请求或任务是安全的；recover 不会修复已经被破坏的共享状态。

这种捕获所有 panic 的策略属于执行边界的设计选择，不是普通业务函数的默认写法。不应在每一层重复恢复并吞掉异常。

## 5. 无法恢复的情况

- 子 goroutine 的 panic：必须在子 goroutine 内设置自己的恢复边界，父 goroutine 的 recover 无效。
- `os.Exit`：直接结束进程，不执行 defer，也不会被 recover 拦截。
- 运行时致命错误：例如并发 map 误用、对未锁定的 Mutex 执行 Unlock，不能依靠 recover 恢复。

recover 也不能终止死锁或解除永久阻塞。它只处理满足调用条件的 panic。

## 6. 实现与源码入口

- `src/builtin/builtin.go`：recover 的文档化声明。
- `src/runtime/panic.go`：panic 展开、延迟调用与恢复的主要运行时逻辑。
- `src/cmd/compile/internal`：defer、panic 和 recover 相关的编译处理。

## 7. 官方参考

- [Effective Go：Recover](https://go.dev/doc/effective_go#recover)：任务边界恢复、包内部 panic 转 error 及意外 panic 重新抛出的示例。
- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)：说明包内部可使用 panic/recover，而对外 API 按约定返回显式错误。
- [encoding/json 编码实现](https://go.dev/src/encoding/json/encode.go)：通过私有 jsonError 类型区分内部有意触发的 panic，其他 panic 重新抛出。
