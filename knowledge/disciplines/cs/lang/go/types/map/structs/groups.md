# Map groups

groups 是 table 中的组数组引用，类型为 groupsReference。它定位该表的连续 group 存储，并保存用于组索引计算的掩码。本文按 Go 1.27.2 官方源码说明。

## 1. 结构字段

类型定义、初始化与单组地址计算参见[groupsReference](groupsReference.md)。

```go
type groupsReference struct {
    data       unsafe.Pointer
    lengthMask uint64
}
```

| 字段 | 作用 |
| --- | --- |
| data | 指向组数组的起始地址 |
| [lengthMask](lengthMask.md) | 组数减一，用于计算组索引 |

组数为 2 的幂，`lengthMask + 1` 即组数。它不是键值条目数，也不是当前已使用的组数。

## 2. 在 table 中的位置

```text
table
    ├─ used、capacity 等表状态
    └─ groups
         ├─ data → [group 0][group 1][group 2]……
         └─ lengthMask：组数 - 1
```

每个 group 包含控制信息及 8 个逻辑键值槽，因此表的总槽位数为 `(lengthMask + 1) × 8`。有效条目数量由 table.used 统计，不保存在 groupsReference 中。

## 3. 如何定位一个组

已知合法组索引 i 时，运行时按类型元数据中的 GroupSize 计算地址：

```text
组地址 = data + i × GroupSize
```

GroupSize 是该 map 类型对应的实际组大小，受键值类型、间接存储、对齐和组内布局等因素影响，不能假设所有 map 的 group 都占相同字节数。

组索引和组内槽索引是不同概念：先定位一个组，再通过控制信息匹配找到组内候选槽。

## 4. lengthMask 与探测

组数为 2 的幂时，可通过与 lengthMask 按位与，将探测得到的位置限制在组数组范围内，代替相应的取模运算。例如组数为 8，lengthMask 就为 7。

查找从哈希确定的起始组出发，按表的探测序列访问其他组，不一定从 group 0 开始依次遍历所有组。控制字节筛选候选槽后，仍需比较完整键。

## 5. group 的内容

每组包含一个 64 位控制字，对应 8 个槽位的控制字节。占用槽保存哈希指纹，另外的特殊值标记空槽和删除槽。

键值数据可以按交错槽位布局，也可以在 mapsplitgroup 实验分支下分离为键数组和值数组。groupsReference 通过统一的组引用和类型元数据访问这些存储，不直接把 data 转成所有类型通用的普通 Go 切片。

## 6. 分配与增长

newGroups 按指定组数及该 map 的组类型准备数组存储，并将 lengthMask 设置为组数减一。组数组在当前表布局中大小固定。

表增长、拆分或重建时，由相应路径重排条目并更新表与目录关系，不能只扩大 lengthMask 而继续沿用原存储。旧组存储是否仍被迭代器等引用，会影响其后续生命周期。

小 map 单组模式则可以由 Map.dirPtr 直接指向 group，不一定存在独立的 table.groups 结构。

## 7. 源码入口

- `src/internal/runtime/maps/group.go`：groupsReference、newGroups、组地址计算和组引用。
- `src/internal/runtime/maps/table.go`：组探测、表容量和重建。
- `src/internal/abi`：组类型、大小、字段偏移及步长等元数据。
