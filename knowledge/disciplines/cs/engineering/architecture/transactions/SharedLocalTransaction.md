# 共享本地事务

共享本地事务指多个模块的数据库操作参加同一个本地事务，全部成功后统一提交；提交前任一步骤失败，则中止并回滚整个事务。模块可以保留私有 store，只需提供让内部数据库操作使用共同事务的机制。

## 1. 前提与执行流程

参与操作必须使用同一个有效事务，涉及的数据表和存储引擎必须支持相应事务能力。仅使用同一个数据库地址、同一个连接池或同一个请求上下文，并不意味着参加了同一事务。

```text
下单用例开启事务
    ↓
订单模块创建订单 ── 私有 store 使用该事务
    ↓
库存模块扣减库存 ── 私有 store 使用该事务
    ↓
全部成功：统一提交
任一步失败：统一回滚
```

库存不足时，订单插入也会回滚，无需再执行取消订单的补偿操作。Saga 的每一步则会先提交自己的本地事务，后续失败时需要业务补偿，两者不同。

## 2. 事务依赖的传递方式

### 2.1 显式参数

需要参与事务的方法接收事务对象或项目定义的事务范围对象，再传给内部 store。模块实例可以在应用启动时创建并长期复用。

```go
// 示意接口：TxScope 封装事务访问能力，生命周期由 uow 管理。
return uow.Run(ctx, func(scope *TxScope) error {
    if err := orders.Create(ctx, scope, orderID); err != nil {
        return err
    }
    return stocks.Deduct(ctx, scope, skuID, quantity)
})
```

依赖清楚，容易检查哪些操作参加事务，但需要调整调用链的方法签名。无需给所有方法增加参数，只有必须参加事务的操作及其内部调用链需要支持。

事务范围不应向模块授予独立提交、回滚的职责；可以只提供查询和执行能力，将生命周期控制保留在事务管理组件中。

### 2.2 通过 context 传递

事务管理组件将事务放入派生的 `context.Context`，模块继续传递该上下文，store 从中获取事务。

```go
return uow.Run(ctx, func(txCtx context.Context) error {
    if err := orders.Create(txCtx, orderID); err != nil {
        return err
    }
    return stocks.Deduct(txCtx, skuID, quantity)
})
```

这种方式保持业务方法签名简洁，但隐藏了事务依赖。传错上下文时，若 store 自动退回普通数据库连接，可能发生部分修改提前提交。

采用该方式时，需要统一事务访问组件和上下文键；必须参加事务的写操作应在缺少事务时报错。对于同时支持事务和非事务调用的方法，需要明确调用契约及检测方式，不能仅依赖约定猜测当前执行范围。

### 2.3 事务范围内创建实例

在事务开始后，通过工厂或依赖注入创建绑定该事务的模块或 store 包装实例：

```go
// 示意接口：modules 是当前事务范围内的依赖集合。
return uow.Run(ctx, func(modules *TxModules) error {
    if err := modules.Orders.Create(ctx, orderID); err != nil {
        return err
    }
    return modules.Stocks.Deduct(ctx, skuID, quantity)
})
```

业务方法不需要事务参数，但需要支持按事务组装依赖。新建的可以只是轻量包装实例，不必重新初始化整个模块，也不必创建新的连接池。

事务范围实例不能在事务结束后继续使用。

## 3. 显式传参示例

下面使用 Go 的 `database/sql` 展示生命周期与事务传递。代码片段省略包声明、导入和模块构造函数；模块在应用启动时创建，内部 store 保持私有。

```go
// 各模块的 store 可以分别在内部声明所需的最小接口。
type executor interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type orderStore struct{}

func (s *orderStore) create(
    ctx context.Context, exec executor, orderID int64,
) error {
    _, err := exec.ExecContext(ctx,
        "INSERT INTO orders (id, status) VALUES (?, ?)",
        orderID, "created",
    )
    return err
}

type OrderModule struct {
    store *orderStore
}

func (m *OrderModule) Create(
    ctx context.Context, tx *sql.Tx, orderID int64,
) error {
    return m.store.create(ctx, tx, orderID)
}

type stockStore struct{}

func (s *stockStore) deduct(
    ctx context.Context, exec executor, skuID int64, quantity int,
) error {
    if quantity <= 0 {
        return errors.New("quantity must be positive")
    }
    result, err := exec.ExecContext(ctx, `
        UPDATE stocks SET available = available - ?
        WHERE sku_id = ? AND available >= ?
    `, quantity, skuID, quantity)
    if err != nil {
        return err
    }
    n, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if n != 1 {
        return errors.New("stock missing or insufficient")
    }
    return nil
}

type StockModule struct {
    store *stockStore
}

func (m *StockModule) Deduct(
    ctx context.Context, tx *sql.Tx, skuID int64, quantity int,
) error {
    return m.store.deduct(ctx, tx, skuID, quantity)
}

func Checkout(
    ctx context.Context, db *sql.DB,
    orders *OrderModule, stocks *StockModule,
    orderID, skuID int64, quantity int,
) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := orders.Create(ctx, tx, orderID); err != nil {
        return err
    }
    if err := stocks.Deduct(ctx, tx, skuID, quantity); err != nil {
        return err
    }
    return tx.Commit()
}
```

示例 SQL 使用 `?` 占位符，要求订单 ID、SKU ID 具有唯一约束。其他数据库驱动可能需要不同的占位符写法。

`defer tx.Rollback()` 在提前返回时尝试回滚；提交成功后调用会返回事务已结束的错误，不会撤销提交。实际项目应按需要记录回滚失败，并处理提交结果不确定的情况。

示例直接传递 `*sql.Tx` 以便理解。更严格的封装可以使用只提供执行能力的事务范围对象，使模块无法自行控制提交与回滚。

## 4. 并发与生命周期约束

- 不要给共享模块设置可变的当前事务字段，例如 `orders.tx = tx`，否则并发请求可能相互覆盖。
- 不要把当前事务交给超出回调生命周期的后台任务；事务提交或回滚后不能继续使用。
- 不要在内部 store 中绕过事务调用普通 `db`，否则该操作不受统一回滚保护。
- 嵌套业务调用应明确复用现有事务、拒绝嵌套或使用保存点；新开一个独立事务不能获得整体原子性。

## 5. 选择与验证

新设计可以优先考虑显式参数，让依赖清楚；已有统一上下文事务组件的项目可以沿用 context 方式。需要保持业务方法签名、且已有依赖注入机制时，可以采用事务范围实例。

验证应覆盖：后续模块失败时前序修改确实回滚、成功时全部提交，以及并发请求不会串用事务。采用 context 方式时，还应验证缺少或传错事务上下文不会悄悄绕过事务。
