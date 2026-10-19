# Map globalDepth

globalDepth 是 Map 中的 uint8 字段，表示常规目录查找使用多少位哈希前缀。以下按 Go 1.27.2 的实现说明。

## 1. 哈希位数与目录索引

常规目录查找使用哈希的最高 globalDepth 位。在常见 64 位平台上，目录索引可以理解为：

```text
globalShift = 64 - globalDepth
目录索引 = hash >> globalShift
```

该版本在 32 位平台及 Wasm 使用 32 位哈希，移位量相应根据 32 位宽度计算。globalDepth 为零时，目录索引专门返回零，不依赖直接移位表达式。

## 2. 与目录长度的关系

常规目录模式下，目录长度为 `1 << globalDepth`，即 2 的 globalDepth 次方：

| globalDepth | 使用的前缀位数 | 目录项数量 |
| --- | --- | --- |
| 0 | 0 | 1 |
| 1 | 1 | 2 |
| 2 | 2 | 4 |
| 3 | 3 | 8 |

例如最高 3 位有 000 至 111 共 8 种组合，各自对应一个目录项。多个目录项可以指向同一个 table，因此这个数量不等于不同 table 的数量。

## 3. 与 localDepth 的区别

Map.globalDepth 作用于整个目录；table.localDepth 表示区分该表需要多少位哈希前缀。localDepth 可以小于 globalDepth，此时一个表对应多个目录项。

例如 globalDepth 为 2、某表 localDepth 为 1，该表通常由两个相邻目录项共同引用。

## 4. 何时增加

表拆分需要再使用一位哈希前缀。如果待拆分表的 localDepth 已等于 globalDepth，现有目录位数不足，就扩展目录，使 globalDepth 加一、目录项数量翻倍，并更新相关指针和移位字段。

如果 localDepth 小于 globalDepth，则现有目录已具备区分新表的空间，不一定需要扩大整个目录。

## 5. 小 map 模式

dirLen 为零时，Map 处于小 map 单组模式，不经过目录选择 table。不能仅凭 globalDepth 为零判断是否为小 map，因为常规目录也可以只有一个目录项。

## 6. 源码入口

- `src/internal/runtime/maps/map.go`：globalDepth、globalShift、目录索引和目录增长。
- `src/internal/runtime/maps/table.go`：表拆分时的局部深度更新。
