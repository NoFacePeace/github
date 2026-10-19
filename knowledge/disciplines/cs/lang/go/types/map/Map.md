# Map

map 是 Go 内建的键值映射类型，写作 `map[K]V`。键必须支持相等比较，值可以是任意类型。map 不保证遍历顺序，也不自动提供并发读写保护。

## 1. 声明与初始化

```go
var empty map[string]int // nil map。
counts := make(map[string]int)
values := map[string]int{"go": 1, "java": 2}
reserved := make(map[string]int, 100)
```

make 的可选大小是初始空间提示，不是长度或容量上限。新 map 的长度为零，之后可以继续添加条目；不能使用 cap 查询 map 容量。

## 2. 读取、写入与删除

从哈希定位、组内筛选到完整键比较的过程参见[Map 查找](lookup/Lookup.md)。

插入与更新的共同流程参见[Map 写入](Write.md)。

清理、表扩容、拆分与目录扩容参见[Map 扩容](growth/Growth.md)。

条目清理、空槽与删除标记的选择过程参见[Map 删除](Delete.md)。

```go
counts := make(map[string]int)
counts["go"] = 1
counts["go"]++

value := counts["missing"] // 不存在时返回值类型的零值，这里为 0。
current, ok := counts["go"] // ok 表示键是否存在。

delete(counts, "go")
clear(counts) // Go 1.21 起提供，删除所有条目。
_ = value
_ = current
_ = ok
```

仅根据零值不能区分键不存在，还是键存在但值恰好为零，需要使用双返回值形式。删除不存在的键不报错，clear 不会把 map 变量变成 nil，也不保证立即向操作系统归还内存。

## 3. 键的要求

| 键类型 | 是否允许 |
| --- | --- |
| 布尔、数值、字符串、指针、channel | 允许 |
| 数组、结构体 | 元素或全部字段类型可比较时允许 |
| 切片、map、函数 | 不允许 |
| 接口 | 静态类型允许，但动态值也必须可比较 |

```go
values := make(map[any]int)
values["go"] = 1
// values[[]int{1}] = 2 // panic：切片不能作为动态键。
```

指针键按指针相等规则比较，不按所指对象的字段内容比较。浮点 NaN 不等于自身，作为键时会出现按同一个 NaN 值也无法正常查回或删除对应条目的情况，应避免这样设计键。

## 4. nil map

nil 是 map 的零值，不等于已初始化的空 map。

| 操作 | nil map 的行为 |
| --- | --- |
| len | 返回 0 |
| 读取合法键 | 返回值类型零值，双返回值中的 ok 为 false |
| range | 不执行循环体 |
| delete、clear | 不产生效果 |
| 写入 | panic，需要先初始化 |

`new(map[K]V)` 返回指向 nil map 的指针，不能替代 make 完成可写 map 的初始化。

## 5. 赋值与函数传参

```go
a := map[string]int{"count": 1}
b := a
b["count"] = 2
// a["count"] 也为 2。
```

赋值和传参复制 map 值，不复制所有键值数据，副本仍关联同一份映射。函数内修改条目可以被调用方看到，但将参数重新赋成另一个 map 不会自动替换调用方的变量。

需要独立映射时，可以创建新 map 并复制条目，或使用标准库 `maps.Clone`。这仍是浅拷贝，指针、切片、map 等值可能继续共享其关联的数据。

## 6. 元素不可寻址

不能取得 `&m[key]`。如果值是结构体，也不能直接用 `m[key].Field = value` 修改其中的普通字段，应先取出副本，修改后再写回：

```go
type User struct {
    Name string
}

users := map[int]User{1: {Name: "Alice"}}
user := users[1]
user.Name = "Bob"
users[1] = user
```

如果 map 保存 `*User`，则可以通过取得的指针修改目标对象，但仍需处理 nil 指针和并发同步。

## 7. 遍历与比较

range 顺序未指定，多次遍历可能不同。需要稳定顺序时，应提取键并排序，再按排序结果访问。

在同一 G 中遍历时，尚未访问的条目如果被删除，就不会再产生；新增条目可能被遍历到，也可能被跳过。这不意味着允许其他 G 无同步地同时修改。

map 只能与 nil 比较，不能直接使用 `==` 比较两个 map 的内容。内容比较可根据值类型使用 `maps.Equal` 或 `maps.EqualFunc` 等工具。

## 8. 并发访问

共享 map 在没有写入时可以由多个 G 并发读取；存在并发写入、删除或清空时，相关读写必须遵守同步协议。

可以使用 Mutex、RWMutex 保护映射，或按场景选择专用的 `sync.Map`。map 的并发误用可能触发运行时致命错误，但不能依赖运行时一定检测到所有错误，也不能依靠 recover 修复这种设计。

对 map 值的重新赋值、对其引用对象的修改，也需要分别考虑同步；只保护查找过程不一定保护返回对象后续的访问。

## 9. 实现与源码入口

Go 1.27.2 的 Map、目录、table 与 group 结构参见[Map 结构](structs/structs.md)。

分组探测、控制信息和哈希指纹的设计参见[瑞士表](../../../../data-structures/hash-table/SwissTable.md)。

Go 的 map 底层使用哈希映射实现，布局、扩容、冲突处理和迭代机制随版本变化。新版实现采用 Swiss Table 相关设计，阅读旧版的桶、溢出桶笔记时需要确认版本。

- `src/internal/runtime/maps`：新版 map 的主要实现。
- `src/runtime` 中 map 相关入口：编译器与运行时辅助接口。
- `src/cmd/compile/internal`：map 操作的类型检查与代码生成。
- `src/maps`：复制、比较和迭代等标准库工具。

make、赋值或扩容不保证每次都发生堆分配，具体取决于实现和编译优化。
