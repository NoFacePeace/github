# Map dirLen

dirLen 是 Map 中的 int 字段，记录 table 指针目录的长度。零值还用于区分小 map 的单组模式。以下按 Go 1.27.2 的实现说明。

## 1. 两种模式

| dirLen | 含义 | dirPtr 的目标 |
| --- | --- | --- |
| 等于 0 | 小 map 单组模式 | 直接指向 group，尚未分配时也可能为 nil |
| 大于 0 | 常规目录模式 | table 指针数组 |

运行时通过 dirLen 判断采用哪个访问路径。小 map 单组最多保存 8 个条目，但当前 `len(m) <= 8` 不代表必然处于单组模式；删除条目不会保证它自动退回小 map。

## 2. 目录项不等于表数量

```text
dirLen = 4
dirPtr → [table A][table A][table B][table B]
```

上述目录有 4 项，但只有 2 个不同的 table。同一个表可以被多个相邻目录项引用，具体数量由全局和局部深度的关系决定。

dirLen 也不是有效条目数，`len(m)` 对应 Map.used，表内槽位总数则由各 table 的容量决定。

## 3. 与 globalDepth 的关系

常规目录模式下：

```go
dirLen = 1 << globalDepth
```

`<<` 是左移运算，1 左移 n 位等于 2 的 n 次方。globalDepth 为 3 时，目录使用 3 位哈希前缀，对应 8 种组合，因此 dirLen 为 8。

这条关系不用于解释小 map 模式的 dirLen 零值。常规模式 globalDepth 为零时，dirLen 为 1，仍有一个目录项。

## 4. 增长与初始化

创建 map 时，大小提示和编译优化会影响是否直接准备常规目录。小 map 增长为常规模式时，需要建立目录并更新 dirLen 等字段；常规目录扩大时，dirLen 通常翻倍。

目录项数变化不等于 map 元素数量同步翻倍。增长过程由运行时维护一致性，业务代码不能单独修改 dirLen 或用它作为并发访问依据。

## 5. 源码入口

- `src/internal/runtime/maps/map.go`：dirLen、目录初始化、单组模式判断及目录增长。
