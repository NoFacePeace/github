# Query Object

Query Object（查询对象）是用对象表示数据库查询的设计模式。Martin Fowler 将其描述为一种解释器：查询由对象结构表达，再转换为 SQL。

调用方通过领域类及其属性表达查询，转换层负责将这些属性映射到数据库表和列，从而将调用方与具体数据库结构隔离。该模式收录于《Patterns of Enterprise Application Architecture》。

## 1. 条件组合示例

以下是 Go 风格的伪代码，只展示表达方式，不代表某个实际库的接口：

```go
condition := And(
    Equal(BookStatus, Published),
    GreaterThan(BookPrice, int64(100)),
)
```

BookStatus 和 BookPrice 表示预定义的领域属性，而不是调用方任意传入的 SQL 列名。转换层根据属性映射和操作符生成参数化 SQL，例如：

```sql
SELECT id, title
FROM books
WHERE status = ? AND price > ?;
```

参数按顺序传入 Published 对应的存储值和 100，不直接拼接进 SQL 文本。需要更复杂的组合时，可用 And、Or、Not 等节点构成条件树，并明确括号和运算优先级。

## 2. 与其他查询方式的区别

| 方式 | 表达方式 | 适用场景 |
| --- | --- | --- |
| 固定查询方法 | `GetByID(ctx, id)`、`ListByStatus(ctx, status)` | 固定、明确的查询意图，参数通常必填。 |
| 简单 Filter | `BookFilter{MinPrice: &minPrice, Statuses: statuses}` | 条件少且已知，通常按固定规则组合。 |
| Query Object | 用属性、操作符和逻辑节点组合查询 | 需要动态条件、复杂组合，并希望隐藏数据库结构。 |

普通 Filter 结构体不自动等于完整的 Query Object。前者可以只是参数容器，由查询构造层显式判断可选值；后者强调以对象表示查询，以及解释、组合和映射查询的机制。

固定查询方法也可以在内部复用 Query Object，两者并不互斥。

## 3. 使用边界

以下是实践上的取舍，不是必须采用的实现形式：

- 简单场景优先使用固定方法或强类型 Filter，不必为了减少几个条件判断就实现通用查询框架。
- 只有确实需要动态组合时，再增加条件树和解释器；条件树也可以采用强类型属性，避免任意字符串字段。
- 校验允许查询的属性、操作符、值类型和组合复杂度。参数化只保护值，不会自动保护动态表名、列名和操作符。
- 租户隔离、权限范围等强制约束由服务端独立施加，不能由调用方提供的条件替代。
- Query Object 负责表达和转换查询，不会自动解决索引选择、性能或分页一致性问题。

## 4. 参考资料

- [Martin Fowler：Query Object](https://martinfowler.com/eaaCatalog/queryObject.html)：模式定义、专用查询方法的局限，以及领域属性到数据库结构的映射思路。
