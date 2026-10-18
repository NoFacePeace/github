# Mutex

`sync.Mutex` 是互斥锁，用于保护共享数据或要求独占执行的操作。同一时刻只有一个持有者，其他调用 Lock 的 G 需要等待。

## 1. 基本用法

```go
import "sync"

type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

零值 Mutex 处于未加锁状态，不需要额外初始化。读取和写入都通过同一个锁保护，避免无锁读取与加锁写入发生数据竞争。

## 2. 加锁与解锁

| 方法 | 行为 |
| --- | --- |
| [Lock](Lock.md) | 取得锁；暂时无法取得时等待 |
| [Unlock](Unlock.md) | 释放锁，按实现需要唤醒等待者 |
| [TryLock](TryLock.md) | 立即尝试取得锁，成功返回 true，失败返回 false |

成功的 TryLock 与 Lock 具有相应的同步效果。失败的 TryLock 不建立同步关系，不能因此无锁读取受保护的数据。TryLock 从 Go 1.18 开始提供。

## 3. 竞争时发生什么

锁状态、信号量字段及等待队列的关联方式参见[Mutex 结构](struct.md)。

信号量字段在加锁和解锁中的作用参见[Mutex sema](sema.md)。

无竞争时，获取锁通常走原子操作的快路径。有竞争时，运行时实现可能按条件短暂自旋，再通过信号量等待等机制挂起 G。持有锁的 G 解锁后，等待者可以被唤醒参与后续执行。

挂起的是 G，M 可以继续调度其他 G；唤醒也不保证立即执行或保证业务侧严格 FIFO。实现中的公平性、自旋和饥饿处理策略以 Go 版本为准。

等待者的唤醒、重复通知抑制及标记清除参见[Mutex 唤醒](Wakeup.md)。

长时间竞争失败时的判断与交接策略参见[Mutex 饥饿](Starvation.md)。

退出饥饿模式的条件和状态更新参见[Mutex 饥饿恢复](StarvationRecovery.md)。

## 4. 内存可见性

Mutex 的解锁与后续成功加锁之间建立同步关系。后续持锁代码可以观察在先前临界区中完成的修改，前提是相关访问遵守同步规则。

临界区应覆盖需要保持一致的完整操作。例如“检查余额并扣款”应作为一个整体保护，不能只分别保护读取和写入而允许中间状态被其他 G 改变。

## 5. 常见问题

- Mutex 不可重入：同一个 G 持锁后再次 Lock，会等待该锁，可能自我死锁。
- 对未锁定的 Mutex 调用 Unlock 会触发不可恢复的运行时致命错误，不能依赖 recover 处理。
- 首次使用后不能复制 Mutex；包含锁的结构体通常使用指针接收者并通过指针传递。
- 锁不绑定某个 G，允许一个 G 加锁、另一个 G 解锁，但必须设计清楚交接协议。
- 多把锁以不同顺序获取，可能形成循环等待；应统一加锁顺序。

## 6. 源码入口

- `src/sync/mutex.go`：对外 API，部分版本委托内部实现。
- `src/internal/sync/mutex.go`：新版实现中的锁状态、自旋与竞争处理。
- `src/runtime/sema.go`：相关等待和唤醒逻辑。

实现位置与细节以所用 Go 版本为准。
