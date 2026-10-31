# 作用域元数据

## 1. 含义与范围

作用域元数据是描述当前请求或任务、需要沿相关调用链传递的附加信息，例如请求 ID、链路追踪信息和调用方身份。这里的作用域不是语法上的 `{}`，而是由 context 派生和传递关系划定的处理范围。

商品 ID、购买数量等核心业务输入通常使用显式参数；不要因为某个值被多个函数使用，就将它放进 `Context`。`WithValue` 也不负责把值自动序列化或发送到另一个进程，跨网络传播需由协议或中间件完成。

## 2. 类型安全的访问

使用包内自定义的可比较 key 类型，避免不同包使用相同字符串 key 产生冲突，并封装类型明确的存取函数。

```go
package requestmeta

import "context"

type ownerKey struct{}

func WithOwner(ctx context.Context, owner string) context.Context {
	return context.WithValue(ctx, ownerKey{}, owner)
}

func OwnerFromContext(ctx context.Context) (string, bool) {
	owner, ok := ctx.Value(ownerKey{}).(string)
	return owner, ok
}
```

这里的 `owner` 只表达当前调用链携带的身份。读取到该值，不代表已经验证身份，也不代表仍持有某个资源。

`Context` 本身支持多个 goroutine 并发调用，但这不保证其中保存的 map、指针或业务对象可以安全地并发修改。

## 3. 与租约信息的边界

租约的持有者、标识和有效期等信息可以视为管理资源的元数据，但“属于元数据”不等于“必须放进 context”。应根据下游职责选择传递方式。

| 下游职责 | 传递建议 |
| --- | --- |
| 日志记录、标识本次调用方 | 可以通过 context 传递 `owner` 等身份信息 |
| 区分具体哪一次租约获取 | 需要本次获取的唯一标识，不能只依赖可复用的 `owner` |
| 续租、释放、校验租约或执行受保护写入 | 优先显式传递租约句柄、校验令牌等必要依赖 |

如果同一个 `owner` 可以反复获取租约，那么旧执行和新执行可能具有相同的 `owner`；仅检查持有者身份不足以区分两次获取。应由实际租约协议定义标识、有效性校验和写入保护，不能只相信 context 中保存的快照。

`context` 的 deadline 是本地取消机制，不等于租约到期时间，也不负责续租。可以在续租失败时取消执行 context，让相关工作协作退出，但取消信号本身不能替代资源端的所有权校验。

## 4. 参考资料

- [context 包文档](https://pkg.go.dev/context)。
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)。
