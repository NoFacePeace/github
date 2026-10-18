# RWMutex

`sync.RWMutex` 是读写锁，允许多个读者同时持有读锁；写锁与其他读锁、写锁互斥。它适合需要保护共享数据且读取操作确实可以并行的场景。

## 1. 基本用法

```go
import "sync"

type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    value, ok := c.data[key]
    return value, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.data == nil {
        c.data = make(map[string]string)
    }
    c.data[key] = value
}
```

RWMutex 的零值可以直接使用。示例还在写入时初始化 map，使 Cache 的零值也能正常使用。

## 2. 互斥关系

| 当前持有状态 | 新读者 RLock | 新写者 Lock |
| --- | --- | --- |
| 无持有者且没有待处理写者 | 可以取得 | 可以取得 |
| 已有读者 | 通常可并发取得，但待处理写者会阻止后续读者进入 | 等待读者退出 |
| 已有写者 | 等待 | 等待 |

写者等待时，后续读者会受到限制，避免写者被持续到来的读者长期阻挡。这不意味着 API 提供严格 FIFO 顺序。

## 3. 读锁只能保护读取

多个 G 可以同时持有读锁，因此不能在 RLock 保护下修改共享数据，包括修改 map、递增计数、延迟初始化缓存等。

如果读取返回指向内部可变数据的指针、切片或 map，解锁后的调用方仍可能与其他访问并发冲突。需要根据业务设计复制数据，或继续使用相应同步协议。

## 4. 不支持升级与降级

持有读锁时直接调用 Lock，不会自动升级成写锁，反而可能等待自己释放读锁而死锁。持有写锁时再调用 RLock，也不能实现自动降级。

递归获取读锁也不安全：如果两次 RLock 之间有写者开始等待，第二次 RLock 可能阻塞，而写者又在等待第一次读锁释放。

如果必须先读后改，可以释放读锁后取得写锁，但必须重新检查条件，因为两次加锁之间共享状态可能已经变化。

## 5. 尝试获取与注意事项

`TryLock` 和 `TryRLock` 从 Go 1.18 开始提供，分别尝试获取写锁和读锁，失败立即返回 false。失败不建立同步关系，也不允许绕过锁访问共享数据。

RWMutex 首次使用后不能复制。对未持有相应状态的锁进行解锁会导致运行时错误，不应将其作为正常控制流程。

RWMutex 不一定比 Mutex 更快。维护读者状态也有成本；临界区很短、竞争不高或写入较多时，Mutex 可能更合适，应结合实际负载测量。

## 6. 源码入口

- `src/sync/rwmutex.go`：读写锁计数、读写等待和唤醒逻辑。
- `src/runtime/sema.go`：相关等待与唤醒支持。
