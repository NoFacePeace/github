# Go 原子操作

标准库 `sync/atomic` 提供原子读取、写入、交换、加法以及比较并交换等操作，用于计数、状态更新和特定同步协议。原子操作保证的是一次操作的不可分割性，不会自动把周围的代码或多个字段组成一个原子事务。

## 1. 常用操作

| 操作 | 类型方法示例 | 作用 |
| --- | --- | --- |
| Load | `value.Load()` | 原子读取当前值 |
| Store | `value.Store(new)` | 原子写入新值 |
| Swap | `value.Swap(new)` | 原子替换，并返回旧值 |
| Add | `value.Add(delta)` | 原子加法，并返回新值 |
| CAS | `value.CompareAndSwap(old, new)` | 等于预期旧值时才替换，返回是否成功 |

并非每种原子类型都有全部操作。例如 Bool 没有 Add，Pointer 的操作也不等同于整数运算。

## 2. 类型化 API

Go 1.19 起提供 `atomic.Int64`、`atomic.Uint64`、`atomic.Bool`、`atomic.Pointer[T]` 等类型，通常比直接传递普通变量地址更清晰。

```go
import "sync/atomic"

type Counter struct {
    value atomic.Int64
}

func (c *Counter) Increment() int64 {
    return c.value.Add(1)
}

func (c *Counter) Value() int64 {
    return c.value.Load()
}
```

这些类型的零值可以直接使用，首次使用后不能复制。包含原子字段的结构体也应避免使用后按值复制。

旧的函数式 API 仍可使用，例如 `atomic.AddInt64(&value, 1)`。某些 32 位平台对这类 64 位操作有对齐要求；类型化 API 可以帮助正确安排对齐。

## 3. CAS 的语义

比较交换的执行过程、重试循环及 ABA 问题参见[CAS](CAS.md)。

CAS 是 Compare-And-Swap，即比较并交换：

```go
var state atomic.Int32

func TryStart() bool {
    return state.CompareAndSwap(0, 1)
}
```

其语义是原子地检查当前值是否等于 0，相等则改成 1 并返回 true，否则不修改并返回 false。如果没有其他操作把状态改回 0，多个 G 同时尝试时只有一个能成功。

普通的“先读再写”不具备同样保证：

```go
// 即使两个操作分别原子，也不能保证整体只被一个 G 成功执行。
if state.Load() == 0 {
    state.Store(1)
}
```

两个 G 都可能先读到 0，再分别写入 1。需要保证比较与更新之间不被冲突修改插入时，应使用 CAS。

## 4. 底层实现

原子操作通常由编译器或底层汇编映射到 CPU 的原子指令，不需要为每次操作进入内核或获取 OS 锁。

| 架构 | CAS 的常见实现 |
| --- | --- |
| x86 / amd64 | `LOCK CMPXCHG` |
| ARM64 | 支持时使用 CAS 指令，或通过独占加载、存储构成重试循环 |

CPU 通过缓存一致性和原子访问机制协调多个核心的竞争。具体指令及辅助实现随目标架构、CPU 能力和 Go 版本变化，不能认为所有操作都只对应一条指令。

## 5. 内存可见性

Go 的原子操作具有顺序一致性语义，可以理解为所有原子操作按某个统一顺序执行。如果一个原子操作的效果被另一个观察到，它们之间建立相应同步关系。

这可以用于正确发布已经准备好的数据，但必须遵守完整的访问协议。通过原子指针发布对象，并不意味着对象字段之后可以任意并发修改；这些修改仍需要锁、原子操作或其他同步。

同一份共享状态的并发访问不能随意混用普通读写和原子读写，也不能仅靠循环读取普通变量等待另一个 G 更新。

## 6. Pointer 与 Value

`atomic.Pointer[T]` 用于原子读写 `*T`，适合发布快照或替换对象指针。零值表示 nil 指针。

`atomic.Value` 可以保存一个具体类型的值，其零值 Load 返回 nil；首次 Store 后，后续保存的值必须保持同一具体类型，且不能 Store 一个 nil 接口。装入 Value 的 map、切片等引用数据不会自动变成不可变对象，后续并发修改仍需同步。

## 7. 使用边界

- 多字段一致性或完整业务操作通常更适合用 Mutex 保护。
- CAS 失败可能需要重试；竞争激烈时，反复重试会增加 CPU 和缓存开销。
- CAS 只检查当前值，无法自动识别值经历了 A → B → A 的变化，即 ABA 问题；相关算法可能需要版本号等额外协议。
- 原子操作不保证业务公平性，也不保证某个 G 在固定时间内重试成功。

“使用原子操作”不等于“整个算法无锁”，也不意味着一定比锁更快，应依据正确性要求和实际负载测量选择。

## 8. 源码入口

- `src/sync/atomic`：对外函数、类型化 API 和 Value。
- `src/internal/runtime/atomic`：运行时内部原子操作及架构相关实现。
- `src/cmd/compile/internal`：原子操作的编译器内建识别与代码生成。

标准库 API 与运行时内部 API 不完全相同，具体实现位置以所用 Go 版本为准。
