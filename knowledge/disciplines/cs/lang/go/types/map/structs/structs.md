# Map 结构

本文按截至 2026 年 10 月 18 日的最新稳定版 Go 1.27.2 官方源码整理。该版本使用 Swiss Table 设计，核心位于 `src/internal/runtime/maps`。以下是内部结构说明，不是业务代码可以依赖的稳定 API。

## 1. 整体层次

```text
map 值 → Map 控制结构
              ├─ 元素数量、哈希种子、目录深度等
              └─ dirPtr → 目录：table 指针数组
                              ↓
                            table
                              ↓
                            groups：组数组
                              ↓
                            group：控制信息 + 8 个键值槽位
```

多个目录项可以指向同一个 table。小 map 还有直接指向单组的优化，不一定经过目录和 table。

## 2. Map 控制结构

顶层控制字段及其职责参见[Map 控制结构](map.md)。

省略源码注释后，Go 1.27.2 的核心结构为：

```go
type Map struct {
    used              uint64
    seed              uintptr
    dirPtr            unsafe.Pointer
    dirLen            int
    globalDepth       uint8
    globalShift       uint8
    writing           uint8
    tombstonePossible bool
    clearSeq          uint64
}
```

| 字段 | 作用 |
| --- | --- |
| used | 全 map 的有效条目数，不包含删除槽；供 len 等逻辑使用 |
| seed | 每个 map 的随机哈希种子 |
| dirPtr | 通常指向 table 指针目录；小 map 时可直接指向单组 |
| dirLen | 目录项数量；小 map 单组模式中为 0 |
| globalDepth | 目录查找使用的哈希位数 |
| globalShift | 为提取目录索引而移出的哈希位数 |
| writing | 写入过程中的检测标记，用于发现部分并发误用，不是互斥锁 |
| tombstonePossible | 是否可能存在删除标记，false 表示已知不存在 |
| clearSeq | clear 调用的序列计数，配合迭代处理清空事件 |

map 值赋值和传参不会复制全部条目，多个副本仍可能关联同一个 Map。nil map 则没有可供写入的已初始化映射。

## 3. 目录与 table

表字段、容量和拆分机制参见[Map table](table.md)。

常规目录长度为 `1 << globalDepth`，使用哈希前缀选择 table。不同 table 可以处于不同的局部深度，多个目录项因而可能共享同一个 table。

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
| used | 该表有效条目数 |
| capacity | 该表总槽位数，为 2 的幂 |
| growthLeft | 触发相应重建处理前剩余的增长额度，受负载和删除标记影响 |
| localDepth | 区分该表所使用的目录哈希前缀位数 |
| index | 该表在目录中的第一个索引；失效旧表可用 -1 标识 |
| groups | 该表的组数组引用 |

该版本单表最大容量为 1024 个槽位。增长可以通过重建某个表，或拆分表并按需扩大目录完成，而不是每次都迁移整个 map 的全部条目；具体路径取决于容量、负载和删除标记。

## 4. groups 与 group

组数组引用定义参见[groupsReference](groupsReference.md)，单组布局与控制信息参见[Map group](group.md)。

组数组的引用结构为：

```go
type groupsReference struct {
    data       unsafe.Pointer
    lengthMask uint64
}
```

data 指向连续的组存储；组数为 2 的幂，lengthMask 等于组数减一，用于按位运算定位组。

每组包含一个 64 位控制字及 8 个逻辑键值槽。实际组类型由编译器结合键和值的类型生成，运行时通过类型元数据描述的偏移和步长访问。

Go 1.27.2 源码包含两种组内布局，受 `GOEXPERIMENT=mapsplitgroup` 影响，以下仅为布局示意：

```text
交错布局：控制字 + [(key, value), (key, value), …]
分离布局：控制字 + [key, key, …] + [value, value, …]
```

因此，不能把键值永远交错排列当作该版本所有构建的固定规则。较大的键或值还可能采用间接存储，具体对齐、偏移和步长应依据类型元数据判断。

## 5. 控制字与哈希指纹

每个槽对应一个控制字节：

| 控制字节 | 含义 |
| --- | --- |
| `0x80` | 空槽 |
| `0xFE` | 删除槽 |
| `0x00` 至 `0x7F` | 已占用，保存 7 位哈希指纹 H2 |

源码将哈希低 7 位提取为 H2，其余部分用于组探测等定位。查找先批量匹配控制信息，筛出候选槽，再比较完整键；指纹相同不代表键相同。

探测必须区分空槽与删除槽，避免删除破坏查找路径。详细算法参见[瑞士表](../../../../../data-structures/hash-table/SwissTable.md)。

## 6. 小 map 优化

单组模式使用 `dirLen == 0`，dirPtr 可以直接指向一个组，省去目录和 table 层。尚未分配组的空 map，dirPtr 也可能为 nil，初始化时机受创建路径与编译优化影响。

该模式最多保存 8 个条目，不需要跨组探测，也不保留删除标记。增长为常规模式后，不能仅因当前 len 降到 8 以下就断定已经变回小 map。

## 7. 与旧版结构的区别

旧版笔记常描述 hmap、bmap、溢出桶和新旧桶迁移。这些字段不能直接套到 Go 1.27.2 的默认实现上。

当前实现的重点是 Map、目录、table、group、控制字和开放寻址探测。写入检测标记也不会使普通 map 自动支持并发读写，业务仍需遵守同步规则。

## 8. 版本与源码

- [Go 官方下载版本列表](https://go.dev/dl/?mode=json)：核实最新稳定版。
- [Go 1.27.2 map.go](https://github.com/golang/go/blob/go1.27.2/src/internal/runtime/maps/map.go)：Map、目录及小 map 路径。
- [Go 1.27.2 table.go](https://github.com/golang/go/blob/go1.27.2/src/internal/runtime/maps/table.go)：table、探测及增长处理。
- [Go 1.27.2 group.go](https://github.com/golang/go/blob/go1.27.2/src/internal/runtime/maps/group.go)：控制字、组引用及两种布局说明。
