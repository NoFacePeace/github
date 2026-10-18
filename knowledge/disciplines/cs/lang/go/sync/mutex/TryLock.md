# Mutex TryLock

`sync.Mutex.TryLock()` 尝试获取互斥锁，成功返回 true，失败立即返回 false，不进入 Lock 的自旋和挂起等待流程。该方法从 Go 1.18 开始提供。

## 1. 基本用法

```go
if !mu.TryLock() {
    return // 本次跳过需要该锁的操作
}
defer mu.Unlock()

// 取得锁后才能访问受保护的数据。
```

只有返回 true 才能按协议执行临界区并在完成后解锁。返回 false 时，没有获得锁，不能继续访问受保护的数据，也不能调用 Unlock。

## 2. 与 Lock 的区别

| 对比 | Lock | TryLock |
| --- | --- | --- |
| 暂时无法取得锁 | 继续尝试，必要时挂起等待 | 返回 false |
| 是否返回成功标志 | 不返回；正常返回即已取得锁 | 返回 bool |
| 是否进入 `lockSlow` | 快路径失败后进入 | 不进入 |
| 获取失败后的处理 | 由锁实现继续协调 | 由调用方决定跳过或采取其他处理 |

这里的“不等待”指不进入锁的等待队列或持续重试，不代表调用过程不会被操作系统或 Go 调度器暂停。

## 3. 实现流程

以下是 Go 1.26.3 内部实现的核心逻辑，省略竞态检测相关代码：

```go
func (m *Mutex) TryLock() bool {
    old := m.state
    if old&(mutexLocked|mutexStarving) != 0 {
        return false
    }
    if !atomic.CompareAndSwapInt32(&m.state, old, old|mutexLocked) {
        return false
    }
    return true
}
```

如果锁已被持有，或处于饥饿模式，直接失败；否则通过 [CAS](../atomic/CAS.md) 尝试设置锁定位。比较交换失败后直接返回，不循环重试。

普通模式下，即使已有等待者，当前运行的 G 仍可能通过 TryLock 取得锁。因此，它不保证排队顺序，也不优先让出机会给已经被唤醒的等待者。具体策略以所用 Go 版本为准。

## 4. 同步关系

成功的 TryLock 与 Lock 一样建立相应同步关系。失败的 TryLock 不建立同步关系，也不能据此判断锁或共享数据此后会保持什么状态。

```go
// 错误思路：失败不表示可以无锁读取共享数据。
// if !mu.TryLock() {
//     return sharedValue
// }
```

如果需要失败时返回缓存或旧快照，该数据必须有自己正确的同步或不可变发布协议。

## 5. 使用场景与限制

TryLock 可用于允许跳过的维护任务，或特定需要避免等待的协调协议。普通共享数据访问通常使用 Lock 更直接，TryLock 的正确使用场景相对少。

不要简单写无限循环反复调用 TryLock 来替代 Lock：这会消耗 CPU，绕过正常等待机制，也不保证公平性。TryLock 不提供超时或 context 取消能力；需要这些语义时，应单独设计相应的协作机制。

同一个 G 已持锁时再次 TryLock 会返回 false，不会因为是原持有者而成功；Mutex 仍然不可重入。

## 6. 源码入口

- `src/sync/mutex.go`：对外 TryLock 方法。
- `src/internal/sync/mutex.go`：新版内部状态检查与比较交换实现。
