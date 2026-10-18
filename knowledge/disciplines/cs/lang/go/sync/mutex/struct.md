# Mutex 结构

Mutex 本身保存锁状态和运行时信号量字段，具体等待的 goroutine 则由运行时等待结构管理，并不直接存放在 Mutex 内。以下按本地 Go 1.26.3 的实现说明，内部布局和队列组织可能随版本变化。

## 1. 对外结构与内部实现

`sync.Mutex` 包含用于辅助检查禁止复制的字段，以及内部 Mutex 实现。其核心同步状态位于 `internal/sync.Mutex`：

```go
type Mutex struct {
    state int32
    sema  uint32
}
```

| 字段 | 作用 |
| --- | --- |
| [state](state.md) | 编码锁定位、唤醒标记、饥饿标记及等待计数 |
| [sema](sema.md) | 运行时信号量计数，其地址还用于关联对应等待者 |

Mutex 不保存持有者 G 的 ID，也没有直接包含一个等待 G 的链表。因此，它不检查调用 Unlock 的 G 是否就是加锁者。

## 2. state 的位布局

```text
高位                         低位
等待计数 … | starving | woken | locked
               bit 2    bit 1   bit 0
```

| 状态 | 含义 |
| --- | --- |
| `mutexLocked` | 当前处于锁定状态 |
| `mutexWoken` | 已有被唤醒或承担竞争的 G，减少重复唤醒 |
| `mutexStarving` | 处于饥饿交接模式，阻止新竞争者抢锁 |
| `state >> mutexWaiterShift` | 锁协议维护的等待者计数，该版本 shift 为 3 |

这些字段通过原子操作协调。不能只检查 locked 位来判断是否可获取：饥饿模式交接期间，即使 locked 暂时为零，新来的 G 也不能抢锁。

## 3. sema 如何关联等待队列

`sema` 不是 OS 信号量句柄，也不是队列指针。运行时用 `&m.sema` 这个内存地址定位等待关系：

```text
Mutex.sema 的地址
    ↓ 哈希定位
全局 semtable 中的 semaRoot
    ↓ 按地址查找
该地址对应的 sudog 等待记录
    ↓
具体等待的 G
```

多个不同地址可以落入同一个桶，但运行时还会按完整地址区分等待者，不会把不同锁的等待队列混为一谈。

因此，首次使用后不能复制 Mutex：除了复制状态可能破坏协议，副本的 sema 地址也与原来的等待关系不同。

## 4. semaRoot 与 sudog

运行时的 `semaRoot` 主要包含内部锁、按等待地址组织的树，以及桶级等待计数。该版本使用 treap 管理不同地址的等待者，同一地址的其他等待记录再通过链式关系组织。

每条等待记录使用 sudog 等结构，关联等待的 G、信号量地址及排队、交接所需状态。sudog 是等待记录，不是新的 goroutine，也不是 OS 线程。

Mutex 的 state 等待计数与 semaRoot 的 nwait 属于不同层次：前者服务于该锁的竞争协议，后者用于运行时桶级等待管理，不能要求它们在任意时刻都相等。

## 5. 登记等待的作用

G 抢不到锁时，锁逻辑先协调 state 等待计数，再进入运行时信号量路径。信号量路径会尝试取得已有令牌，必要时再登记具体 G：

```text
更新 Mutex 的等待状态
    ↓
进入信号量等待路径，尝试取得令牌
    ├─ 成功 → 不必挂起
    └─ 未成功
         → 在内部锁保护下增加等待计数并再次检查
         → 仍需等待时，将 sudog 加入对应地址的队列
         → 提交挂起 G，同时释放内部锁
```

解锁方释放信号量时，增加令牌，并按需从对应地址的等待队列中选择 G 唤醒。令牌、再次检查和队列锁共同协调登记与通知的交错，避免“解锁已经通知，而等待者刚好睡下”导致丢失唤醒。

## 6. 源码入口

- `src/sync/mutex.go`：对外 Mutex 结构与方法。
- `src/internal/sync/mutex.go`：state、sema 与状态位定义。
- `src/runtime/sema.go`：semtable、semaRoot、等待队列与信号量操作。
- `src/runtime/runtime2.go`：sudog 等运行时结构。
