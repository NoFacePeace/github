# groupsReference data

data 是 groupsReference 中的 unsafe.Pointer 字段，指向组数组的起始地址。本文按 Go 1.27.2 的 `src/internal/runtime/maps/group.go` 说明。

## 1. 指向组数组

```text
groupsReference
    ├─ data → [group 0][group 1][group 2]……
    └─ lengthMask：组数减一
```

data 不保存组数，也不直接表示某个键值的位置。组数由 lengthMask 配合内部约定确定；每组大小及组内访问方式由 map 的类型元数据描述。

单组的控制信息、键值布局及引用方式参见[Map group](group.md)。

## 2. 如何初始化

newGroups 根据组类型和组数准备数组存储，再把起始地址写入 data：

```go
data: newarray(typ.Group, int(length))
```

这是运行时内部调用，typ.Group 描述当前 map 的实际组类型。不能把 data 当成所有 map 都通用的普通 Go 切片。

## 3. 定位单个组

groupsReference.group 使用合法组索引 i 和 GroupSize 计算目标地址：

```go
offset := uintptr(i) * typ.GroupSize
return groupReference{
    data: unsafe.Pointer(uintptr(g.data) + offset),
}
```

这里 g.data 是数组起点，返回的 groupReference.data 则定位其中一个组。两个字段都叫 data，但所属结构和指向范围不同；i 为零时，计算得到的地址与起点相同。

地址计算本身不提供普通切片的下标检查，调用方需要保证索引在组数组范围内。

## 4. 存储与生命周期

复制 data 指针不会复制底层组数组。表重建、拆分或替换时，运行时按协议调整相关引用，旧数据的生命周期还可能受迭代器等引用影响。

修改掩码不会自动扩大 data 指向的存储。业务代码不应通过 unsafe 自行修改这些内部字段，也不能依靠它们实现普通 map 的并发读写。

## 5. 源码入口

- `src/internal/runtime/maps/group.go`：groupsReference.data、newGroups 和 group 方法。
- `src/internal/abi`：GroupSize 与组布局元数据。
