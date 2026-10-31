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

### 4.2 内部恢复需要明确的接口约定

当任务函数能够妥善处理 panic，或接口明确承诺将内部 panic 转成 error 时，可以在内部恢复。使用命名返回值保存错误，避免恢复后返回零值 nil，让调用方误以为成功：

```go
func doTask(work func() error) (err error) {
    defer func() {
        if v := recover(); v != nil {
            log.Printf("doTask panic: %v\n%s", v, debug.Stack())
            if cause, ok := v.(error); ok {
                err = fmt.Errorf("doTask panic: %w", cause)
            } else {
                err = fmt.Errorf("doTask panic: %v", v)
            }
        }
    }()

    return work()
}
```

示例需要导入 `fmt`、`log` 和 `runtime/debug`。转换为 error 后，调用方仍需处理失败；不应在每一层重复恢复并吞掉异常。

## 5. 无法恢复的情况

- 子 goroutine 的 panic：必须在子 goroutine 内设置自己的恢复边界，父 goroutine 的 recover 无效。
- `os.Exit`：直接结束进程，不执行 defer，也不会被 recover 拦截。
- 运行时致命错误：例如并发 map 误用、对未锁定的 Mutex 执行 Unlock，不能依靠 recover 恢复。

recover 也不能终止死锁或解除永久阻塞。它只处理满足调用条件的 panic。

## 6. 实现与源码入口

- `src/builtin/builtin.go`：recover 的文档化声明。
- `src/runtime/panic.go`：panic 展开、延迟调用与恢复的主要运行时逻辑。
- `src/cmd/compile/internal`：defer、panic 和 recover 相关的编译处理。
