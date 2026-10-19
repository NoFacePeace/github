# Map table

table 是 Go map 中管理一组 Swiss Table 槽位的内部结构。Map 的目录根据哈希前缀选择 table，再由表内探测定位具体的组和键值。本文按 Go 1.27.2 官方源码说明。

## 1. 结构字段

```go
type table struct {
    used       uint16
    capacity   uint16
    growthLeft uint16
    localDepth uint8
    index      int
    groups     groupsReference
}
```

| 字段 | 作用 |
| --- | --- |
| used | 当前表的有效条目数，不包含删除槽 |
| [capacity](capacity.md) | 总槽位数，为 2 的幂 |
| [growthLeft](growthLeft.md) | 在负载与删除标记约束下，剩余的增长额度 |
| localDepth | 目录区分该表所需的哈希前缀位数 |
| index | 指向该表的第一个目录项索引；旧表失效后可以为 -1 |
| [groups](groups.md) | 保存控制信息与键值数据的组数组引用 |

capacity 是内部槽位数量，不是语言层面可通过 cap 查询的 map 容量。

## 2. table 与目录

Map.globalDepth 描述整个目录使用多少位哈希前缀，table.localDepth 描述区分当前表需要多少位。localDepth 可以小于 globalDepth，因此多个目录项可以指向同一个表。

```text
目录： [A][A][B][B]
        ↓     ↓
      table A table B
```

在该组织方式下，一个表通常对应 `1 << (globalDepth - localDepth)` 个目录项。index 记录第一个位置，辅助替换或拆分时更新相关目录项。

## 3. table 与 groups

groups 指向组数组，每组有 8 个逻辑键值槽及对应控制信息。组数为 2 的幂，总容量满足：

```text
capacity = 组数 × 8
```

表内查询从哈希确定的起始组开始，按探测序列检查控制字节，通过指纹筛选候选槽，再比较完整键。哈希冲突通过表内探测处理，不为每个桶挂键值链表。

组数组在该表的一次有效布局中大小固定；需要重建或扩容时，通过相应路径准备替代存储并调整条目位置。

## 4. used 与 growthLeft

used 统计有效键值条目；growthLeft 用于控制负载和重建时机，不应简单理解为 `capacity - used`。

常规表的平均负载限制通常按每组最多约 7 个有效或相关占用槽位的预算管理，并保留查找所需的空槽；最小单组表存在专门规则。删除标记也会影响增长预算，即使 used 不高，探测仍可能因删除标记而变长。

复用删除槽、删除时能否标为空槽、清理或重建等操作，会按协议维护预算。业务不能依据某个固定负载比例推断每次插入的精确分配行为。

## 5. 增长与拆分

Go 1.27.2 的 `maxTableCapacity` 为 1024 个槽位。需要增长时，表通常在上限内增加容量并重排条目；达到表容量上限后，则可以拆分为两个表，增加区分它们所需的哈希前缀位数。

```text
表内需要更多空间
    ├─ 尚可增长 → 创建更大替代表，重新安排条目
    └─ 达到单表上限 → 拆分成两个表
                           ↓
                      更新目录中的相关指针
```

拆分时，如果当前 localDepth 已等于 globalDepth，目录还需要先扩展以提供更多哈希前缀位；否则可以在已有目录中区分两个新表。

这种设计将主要的条目重排限制在单个表范围，但目录扩展仍可能涉及目录指针数组的复制，不能理解为任何增长操作都没有与目录大小相关的成本。

## 6. 与小 map 的区别

小 map 的单组模式可以直接让 Map.dirPtr 指向 group，不经过独立 table。这里的 table 结构主要属于常规目录模式，不能认为每个 map 从创建起都已分配 table 和完整组数组。

## 7. 旧表与并发

表被替换后，迭代器可能仍需保留旧表来维护遍历语义。旧表的有效生命周期由运行时管理，不能看到 index 为 -1 就直接推断其存储已释放。

目录和表增长的内部协调不等于业务并发安全，普通 map 的冲突读写仍需要同步。

## 8. 源码入口

- `src/internal/runtime/maps/table.go`：table、探测、增长和拆分。
- `src/internal/runtime/maps/map.go`：目录索引、目录替换与分裂更新。
- `src/internal/runtime/maps/group.go`：groupsReference、组存储及控制信息访问。

上述字段和常量对应 Go 1.27.2，后续版本可能调整。
