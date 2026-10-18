# Mutex 解锁

`sync.Mutex.Unlock()` 释放互斥锁，并根据当前状态决定是否唤醒或交接给等待者。本文按本地 Go 1.26.3 的 `src/internal/sync/mutex.go` 整理，内部策略以所用版本为准。

## 1. 入口与快路径

Unlock 先通过原子减法清除锁定位，以下省略竞态检测相关代码：

```go
func (m *Mutex) Unlock() {
    new := atomic.AddInt32(&m.state, -mutexLocked)
    if new != 0 {
        m.unlockSlow(new)
    }
}
```

减去 `mutexLocked` 不是清空整个状态，而是保留等待计数等信息。新状态为零时，没有其他状态需要处理，可以直接返回；否则进入 `unlockSlow`。

## 2. 检查解锁是否合法

慢路径先检查原锁定状态。如果对未锁定的 Mutex 调用 Unlock，会触发 `sync: unlock of unlocked mutex` 致命错误，不能依赖 recover 恢复。

Mutex 不记录持有者 G 的身份，因此允许一个 G 加锁、另一个 G 解锁，但调用方必须保证交接和数据访问协议正确。

## 3. 普通模式下唤醒

普通模式中，解锁方检查是否需要增加一个参与竞争的等待者。以下情况可以直接返回：

- 没有等待者。
- 已经有其他 G 取得锁。
- 已经设置 `mutexWoken`，无需重复唤醒。
- 观察到锁进入饥饿模式，由对应交接路径负责处理。

需要唤醒时，通过 CAS 减少锁状态中的等待计数，并设置 `mutexWoken`；成功后调用 `runtime_Semrelease(&m.sema, false, ...)`。

被唤醒的 G 不自动拥有锁，恢复后仍可能输给新到来的竞争者。Unlock 返回也不代表目标 G 已经开始执行。

## 4. 饥饿模式下交接

饥饿模式下，解锁路径调用 `runtime_Semrelease(&m.sema, true, ...)`，请求将机会交给等待者，并在运行时条件允许时让出执行机会。

此时锁定位可能暂时没有设置，但 `mutexStarving` 仍阻止新来的 G 抢锁。接收交接的等待者恢复后，补齐锁定位、等待计数和模式状态。

因此，不能只看 `mutexLocked` 一位是否为零，就认为锁可以被任意 G 获取。

## 5. 完整流程

```text
Unlock
    ↓
原子清除锁定位
    ├─ 新状态为零 → 返回
    └─ unlockSlow
         → 检查是否对未锁定的锁解锁
         ├─ 普通模式
         │    ├─ 不需要唤醒 → 返回
         │    └─ CAS 取得唤醒资格 → Semrelease，唤醒竞争者
         └─ 饥饿模式 → Semrelease，请求交接给等待者
```

运行时信号量处理 G 的等待与唤醒，并不意味着每次 Unlock 都执行 OS 线程唤醒系统调用。是否需要唤醒或创建工作 M，由后续调度状态决定。

## 6. 内存可见性

一次 Unlock 与后续成功 Lock 之间建立同步关系，使后续临界区能够观察此前受保护的修改。所有相关访问仍需遵守同一同步协议，不能只保护写入而绕过锁读取。

常见写法是取得锁后立即安排 `defer mu.Unlock()`，减少提前返回时遗漏解锁的风险；应同时控制临界区范围，避免无必要地长时间持锁。

## 7. 源码入口

- `src/sync/mutex.go`：对外 Unlock 方法。
- `src/internal/sync/mutex.go`：Unlock 快路径与 `unlockSlow`。
- `src/runtime/sema.go`：信号量释放、等待者唤醒及交接支持。
