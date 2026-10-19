# groupsReference lengthMask

lengthMask 是 groupsReference 中的 uint64 字段，值等于组数减一，用于将探测位置限制在组数组索引范围内。本文按 Go 1.27.2 官方源码说明。

## 1. 为什么是组数减一

常规表的组数为 2 的幂。若有 `2^k` 个组，组索引需要 k 位，而组数减一的二进制恰好是低 k 位全为 1。

| 组数 | lengthMask | 二进制掩码 | 合法组索引 |
| --- | --- | --- | --- |
| 1 | 0 | `0` | 0 |
| 2 | 1 | `1` | 0～1 |
| 4 | 3 | `11` | 0～3 |
| 8 | 7 | `111` | 0～7 |
| 16 | 15 | `1111` | 0～15 |

因此，`position & lengthMask` 保留所需的低 k 位，结果等价于对组数取模。该关系依赖组数为 2 的幂，不能随意套用于其他组数。

## 2. 定位起始组

运行时将哈希右移 7 位得到 H1，再计算：

```go
offset := uint64(hash >> 7) & lengthMask
```

例如 8 个组时，掩码为 111，实际取原哈希第 7～9 位作为起始组索引，位编号从最低位 0 开始。原哈希低 7 位用于 H2 指纹，不参与这次组索引计算。

lengthMask 决定取多少位，哈希决定这些位的值；它不是所有 map 通用的固定常量，而是随当前表的组数变化。

## 3. 后续探测

初始组未命中且需要继续查找时，probeSeq.next 更新位置：

```go
s.index++
s.offset = (s.offset + s.index) & s.mask
```

这里 mask 来自 lengthMask。递增偏移经掩码处理后回绕到组数组范围内，不会因为位置数值增加就直接越过数组末端。

探测中的按位运算负责生成合法索引；实际组地址由数据起点、索引及 GroupSize 计算，二者是不同步骤。

## 4. 与长度、容量的关系

```text
组数 = lengthMask + 1
table.capacity = (lengthMask + 1) × 8
```

每组包含 8 个逻辑槽位，但不保证所有槽位都能长期填满而无需增长。负载预算由表级协议管理，不能从 lengthMask 推断有效条目数、删除槽数量或剩余增长额度。

groupsReference 没有普通切片的独立 len、cap 字段，也不能把 lengthMask 当作切片容量使用。

## 5. 初始化与更新

newGroups 根据 length 分配组数组，并设置 `lengthMask = length - 1`。表增长、重建或拆分时，由运行时维护组存储与掩码的一致性。

仅修改 lengthMask 不会分配更多组，也不能扩大原有存储的合法访问范围。未初始化结构中字段为零，不代表已经存在一个可访问的组。

## 6. 源码入口

- `src/internal/runtime/maps/group.go`：groupsReference、newGroups 和组地址访问。
- `src/internal/runtime/maps/table.go`：makeProbeSeq、probeSeq.next 及表内探测。
