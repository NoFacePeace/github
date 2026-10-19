# Map 定位 group

常规目录查找选定 table 后，运行时使用哈希的 H1 部分计算表内起始组索引，再通过组数组地址定位 group。本文按 Go 1.27.2 官方源码说明。

## 1. 拆分哈希用途

```go
H1 := hash >> 7
H2 := hash & 0x7f
```

H1 用于生成组探测位置，H2 是后续匹配控制字节的 7 位指纹。这里继续使用前面计算的同一份哈希，不是针对 table 再计算一个新的键哈希。

目录选表使用哈希高位；起始组定位则通过掩码取 H1 的相应低位，因此两步使用哈希的方式不同。

## 2. 计算起始组索引

表内组数是 2 的幂，groups.lengthMask 等于组数减一。源码通过 makeProbeSeq 初始化探测状态：

```go
seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
```

初始 offset 的计算为：

```go
offset := uint64(H1) & lengthMask
```

例如有 8 个组，lengthMask 为二进制 111；H1 的低 3 位为 101 时，起始组索引就是 5。只有一个组时掩码为零，起始索引总是零。

组数组引用的字段含义参见[groupsReference](../structs/groupsReference.md)。

## 3. 根据索引计算组地址

运行时调用 `t.groups.group(typ, seq.offset)`，根据 GroupSize 计算地址：

```text
组地址 = groups.data + 组索引 × GroupSize
```

groups.data 指向组数组起点，GroupSize 是该 map 类型对应的实际组大小，计算结果包装为 groupReference。

数据起点及地址计算参见[groupsReference data](../structs/data.md)，目标组的布局参见[Map group](../structs/group.md)。

## 4. 未命中时如何定位下一组

该序列属于二次探测中的[三角数探测](../../../../../data-structures/hash-table/TriangularProbing.md)。

定位起始组后，先检查组内 H2 候选并比较完整键。没有找到且没有可终止查询的空槽时，使用 seq.next 更新探测位置：

```go
s.index++
s.offset = (s.offset + s.index) & s.mask
```

偏移依次累加 1、2、3 等增量，再通过掩码限制范围，而不是始终逐组加一。

例如组数为 8、起始索引为 5，完整一轮组索引为：

```text
5 → 6 → 0 → 3 → 7 → 4 → 2 → 1
```

查找按探测条件及时停止，不必每次访问整轮。定位到一个组只表示确定检查位置，不代表目标键存在。

## 5. 与槽位定位的区别

组内指纹筛选、候选槽访问与完整键比较参见[Map 定位槽位](LocateSlot.md)。

组索引选择表中的一个 group；槽索引则选择该组内 8 个逻辑键值位置之一。后续通过控制字节匹配得到候选槽，再比较完整键确认命中。

不能直接用组索引访问某个键，也不能把 H2 数值当作槽位下标。

## 6. 小 map 单组模式

小 map 单组模式由 Map.dirPtr 直接提供组地址，不经过 table.groups 的起始索引计算，也不执行跨组探测。特定键类型还可能有专用查找优化。

## 7. 源码入口

- `src/internal/runtime/maps/map.go`：h1、h2 与目录模式选择。
- `src/internal/runtime/maps/table.go`：makeProbeSeq、probeSeq.next 和表内查找循环。
- `src/internal/runtime/maps/group.go`：groupsReference.group 与组地址计算。
