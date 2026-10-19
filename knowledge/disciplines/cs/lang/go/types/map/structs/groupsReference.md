# groupsReference

groupsReference 是 Go map 中保存组数组引用的内部结构，table.groups 的类型就是它。本文按 Go 1.27.2 的 `src/internal/runtime/maps/group.go` 说明。

## 1. 结构定义

```go
type groupsReference struct {
    data       unsafe.Pointer
    lengthMask uint64
}
```

它不是普通 Go 切片，没有独立的 len 和 cap 字段。运行时根据约定的组数和类型元数据访问底层组数组。

## 2. data

组数组起始地址的含义和单组定位过程参见[groupsReference data](data.md)。

data 指向组数组的起始地址，概念上可理解为指向若干个对应 map 组类型的连续存储。

每组的实际大小由类型元数据中的 GroupSize 描述，随键值类型、布局与对齐等因素变化。不能把所有 map 的 data 都解释为同一种固定大小的结构数组。

## 3. lengthMask

lengthMask 等于组数减一。组数必须为 2 的幂，因此探测过程中可用按位与计算范围内的组索引。

| 组数 | lengthMask | 逻辑槽位数 |
| --- | --- | --- |
| 1 | 0 | 8 |
| 2 | 1 | 16 |
| 8 | 7 | 64 |

lengthMask 不统计有效条目或删除槽，也不单独表示增长预算。这些状态由 table 的其他字段维护。

## 4. 创建组数组引用

newGroups 为指定组数分配存储，并初始化描述信息：

```go
func newGroups(typ *abi.MapType, length uint64) groupsReference {
    return groupsReference{
        data:       newarray(typ.Group, int(length)),
        lengthMask: length - 1,
    }
}
```

这是运行时内部代码，调用方需要保证 length 满足组数约束。不能因为结构字段的零值是零，就把一个未初始化的 groupsReference 当作已经分配好一个组。

## 5. 从数组引用定位单组

group 方法根据组索引 i 计算地址，再返回单组引用 groupReference：

```go
offset := uintptr(i) * typ.GroupSize
return groupReference{
    data: unsafe.Pointer(uintptr(g.data) + offset),
}
```

该方法本身不通过 lengthMask 自动校正索引，也不提供普通切片下标的边界检查。合法索引由调用方和探测协议保证；按掩码生成索引与按索引计算地址是两个步骤。

groupsReference 管理组数组，groupReference 则定位其中一个组，二者不能混淆。

## 6. 复制与生命周期

复制 groupsReference 只复制指针和掩码，不复制组内键值数据。表重建或拆分时，运行时按协议准备和替换存储，并维护迭代等路径需要的旧引用。

单独修改 lengthMask 不会自动扩大底层数组，也不能让原本越界的访问变得合法。上述结构仅供理解源码，业务代码不应通过 unsafe 直接操作。

## 7. 源码入口

- `src/internal/runtime/maps/group.go`：groupsReference、newGroups 和 group 方法。
- `src/internal/runtime/maps/table.go`：table.groups 的使用与组探测。
- `src/internal/abi`：map 组类型和布局元数据。
