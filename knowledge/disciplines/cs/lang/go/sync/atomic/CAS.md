# CAS

CAS 是 Compare-And-Swap，即比较并交换。它将“比较当前值与预期值”和“相等时写入新值”作为一次原子操作完成，常用于竞争状态切换、更新共享计数或实现同步算法。

## 1. 操作语义

```go
var state atomic.Int32
success := state.CompareAndSwap(0, 1)
```

示例需要导入 `sync/atomic`。CAS 的逻辑可理解为：

```text
原子地执行：
    如果当前值等于 old：
        写入 new，返回 true
    否则：
        不修改，返回 false
```

不能将它替换为普通的 if 判断和赋值；即使分别使用原子 Load 和 Store，两次操作之间仍可能插入其他 G 的修改。

## 2. 多个 G 竞争

假设初始状态为 0，多个 G 同时执行 `CompareAndSwap(0, 1)`，期间没有操作把状态改回 0：

```text
G1：CAS(0, 1) → 成功，状态变成 1
G2：CAS(0, 1) → 失败，状态保持 1
G3：CAS(0, 1) → 失败，状态保持 1
```

只有一个调用能完成从 0 到 1 的转换，具体是谁不保证。失败只说明比较交换发生时当前值不等于预期值，不表示函数调用发生异常。

## 3. 底层实现

CAS 通常通过 CPU 的原子指令或原子指令序列实现，由 Go 编译器和架构相关代码提供支持。

| 架构 | 常见实现 |
| --- | --- |
| x86 / amd64 | 带 LOCK 前缀的 CMPXCHG |
| ARM64 | 支持时使用 CAS 指令，或通过独占加载、存储组成重试循环 |

硬件协调多个核心对目标内存的访问，保证比较与条件写入不会被冲突修改分割。通常不需要进入内核，也不需要先取得操作系统锁；具体实现依目标架构、CPU 能力和 Go 版本而定。

## 4. CAS 重试循环

当新值依赖旧值时，可以先读取、计算，再尝试交换；失败则重新读取并计算：

```go
func Increment(value *atomic.Int64) int64 {
    for {
        old := value.Load()
        next := old + 1
        if value.CompareAndSwap(old, next) {
            return next
        }
    }
}
```

此示例用于说明重试机制，单纯加一直接使用 `value.Add(1)` 更清楚。失败后不能一直使用同一个过期 old，否则可能无效重试。

重试体中的计算可能执行多次，因此不应随意放入发送消息、写文件等不可重复的外部副作用。竞争激烈时，循环还可能持续消耗 CPU，CAS 本身不负责挂起 G 或保证公平性。

## 5. 与互斥锁的关系

Mutex 的无竞争加锁路径可以用 CAS 将状态从未锁定改为锁定；失败后再进入自旋、等待和唤醒等竞争处理。

CAS 只负责一次状态更新，Mutex 则通过状态协议和运行时等待机制保护整个临界区。多字段一致性、复杂业务事务通常更适合由锁保护，不能因为某个标志使用 CAS 就认为全部相关访问都已安全。

## 6. ABA 问题

CAS 只比较当前值，无法直接识别其变化历史：

```text
G1 读取 A
    ↓
G2 将 A 改成 B，再改回 A
    ↓
G1 执行 CAS(A, C)，比较仍可能成功
```

如果算法需要确认中间没有发生变化，仅比较 A 不足以保证正确性，可能需要版本号、不可变对象及配套生命周期协议。版本号与数据也必须按设计协调更新，不能各自随意修改。

## 7. 内存可见性与限制

Go 的原子操作具有顺序一致性语义。正确的原子同步协议可以发布数据，但不会自动保护发布后继续修改的普通字段。

同一共享值的并发访问应遵守一致的同步规则。CAS 不会自动保证周围的普通读写安全，也不保证每个竞争者能在固定时间内成功。

## 8. 常用 API

| API | 对象 |
| --- | --- |
| `atomic.Int32.CompareAndSwap`、`atomic.Int64.CompareAndSwap` | 有符号整数 |
| `atomic.Uint32.CompareAndSwap`、`atomic.Uint64.CompareAndSwap`、`atomic.Uintptr.CompareAndSwap` | 无符号整数 |
| `atomic.Bool.CompareAndSwap` | 布尔状态 |
| `atomic.Pointer[T].CompareAndSwap` | 类型化指针 |
| `atomic.CompareAndSwapInt32` 等 | 函数式 API，接收目标地址 |

类型化原子对象首次使用后不能复制。API 的支持版本和架构约束应以实际 Go 版本为准。
