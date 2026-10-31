# Context

`context` 是 Go 标准库包，`context.Context` 是其中的接口类型，用于沿调用链传递取消信号、截止时间和请求范围的数据。它不是业务参数容器，也不会强制终止 goroutine。

## 1. 接口能力

| 方法 | 含义 |
| --- | --- |
| `Deadline()` | 返回截止时间，以及是否设置了截止时间 |
| `Done()` | 返回取消通知 channel；取消时关闭，永远不会取消的 context 可以返回 nil |
| `Err()` | 未取消时返回 nil；取消后返回 `context.Canceled` 或 `context.DeadlineExceeded` |
| `Value(key)` | 查找当前 context 或祖先 context 中关联的值，未找到时返回 nil |

`Done()` 关闭表示应停止相关工作，不表示工作已经退出。等待 goroutine 完成仍需额外的同步机制。

## 2. 创建与派生

| API | 用途 |
| --- | --- |
| `context.Background()` | 创建没有取消信号、截止时间和值的根 context |
| `context.TODO()` | 暂时无法确定应使用哪个 context 时的占位选择 |
| `context.WithCancel(parent)` | 派生可以主动取消的 context，并返回取消函数 |
| `context.WithDeadline(parent, deadline)` | 派生具有指定截止时间的 context，并返回取消函数 |
| `context.WithTimeout(parent, timeout)` | 按相对时长设置截止时间，并返回取消函数 |
| `context.WithValue(parent, key, value)` | 派生携带一个键值关联的 context |

派生不会修改父 context。子 context 可以读取祖先的值；相同 key 在子 context 中设置新值时，只遮蔽该分支中的查找结果。

对于上述保留父级取消关系的派生方式，父 context 取消会向下传播；取消子 context 不会取消父 context 或兄弟分支。子 context 也不能通过设置更晚的截止时间延长父级允许的执行时间。

## 3. 取消与超时

获取取消函数后，应在适当的生命周期边界调用它，通常紧接创建语句使用 `defer cancel()`，及时释放相关资源。即使预期超时会自动发生，也不应省略取消函数。

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	fmt.Println(wait(ctx, time.Second))
}
```

该例通过 `select` 协作响应取消。实际调用数据库、HTTP 客户端等支持 context 的接口时，也应继续传入当前 `ctx`。只创建带超时的 context，不会自动使不检查取消信号的计算或阻塞操作退出。

调用 `cancel()` 不会等待工作退出；事务回滚、租约释放等清理是否自动发生，取决于所调用组件的协议，不能仅凭取消信号判断清理完成。

## 4. 作用域元数据

- [作用域元数据](ScopedMetadata.md)：请求或任务范围的数据、类型安全的存取方式，以及与业务参数和租约信息的边界。

## 5. 传递约定

- 将 `ctx context.Context` 作为需要它的函数的第一个参数，业务参数另行传递。
- 不传 nil；在确实需要根 context 时使用 `Background()`，尚未明确传递关系时使用 `TODO()`。
- 一般不要将 context 长期保存在结构体中，应按每次操作的生命周期传入。
- 调用链已有 context 时，继续传递或派生，不要随意改用 `Background()` 丢失上游取消和截止时间。
- 不把 `Value` 当作可选参数或通用依赖注入容器；错误处理也不应仅靠 context value。

## 6. 参考资料

- [context 包文档](https://pkg.go.dev/context)。
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)。
- [Contexts and structs](https://go.dev/blog/context-and-structs)。
