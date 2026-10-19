# Map dirPtr

`dirPtr` 是 Go map 内部 Map 结构中的 unsafe.Pointer 字段。它通常指向 table 指针目录，小 map 模式下则可以直接指向单个 group。本文按 Go 1.27.2 官方源码说明。

## 1. 常规模式：指向目录

目录目标表的字段、组数组及增长机制参见[Map table](table.md)。

`dirLen > 0` 时，dirPtr 指向目录数组，概念上的类型为 `*[dirLen]*table`：

```text
Map.dirPtr
    ↓
目录：[table A 指针][table A 指针][table B 指针][table B 指针]
            ↓                         ↓
         table A                   table B
            ↓                         ↓
         groups                    groups
```

目录存放的是 table 指针，不是直接存放每个键值，也不是每个目录项都必须对应不同的表。多个相邻目录项可以指向同一个 table。

## 2. 如何选择目录项

运行时计算键的哈希值，并根据 globalDepth 和 globalShift 提取哈希前缀，得到目录索引。常规目录长度为 `1 << globalDepth`。

```text
键与 Map.seed → 哈希值
    ↓ 提取目录索引
dirPtr 指向的目录项
    ↓
目标 table → 组探测 → 候选槽位 → 比较完整键
```

目录只负责选择 table，进入表后仍需根据探测和指纹规则查找实际条目，不能将目录索引直接当成最终元素位置。

## 3. 小 map 模式：直接指向组

`dirLen == 0` 时采用单组模式，dirPtr 可以直接指向一个 group：

```text
Map.dirPtr → group：控制字 + 8 个逻辑键值槽
```

这种模式跳过目录和 table 层。尚未分配组的空 map，其 dirPtr 也可能为 nil，首次写入等路径再准备存储；编译器也可能提前提供组存储。

单组模式不能只根据当前 len 判断：一个已经增长为常规模式的 map，不会仅因为删除后剩余条目较少就必然回到单组模式。

## 4. 为什么使用 unsafe.Pointer

同一个字段需要保存目录地址或组地址，源码使用 unsafe.Pointer 表示，再结合 dirLen 等内部协议解释目标布局。这是运行时实现手段，不意味着业务代码可以安全地自行转换或修改它。

dirPtr 为 nil 也不能单独判断语言层面的 map 是否为 nil。已初始化但尚未分配组的空 map，可以具有非 nil 的 Map 控制结构和 nil 的 dirPtr，之后仍能正常写入。

## 5. 增长时的变化

小 map 增长后，需要转为常规表组织，dirPtr 的含义随内部状态一起调整。常规模式中，表重建或拆分会更新目录项；目录需要扩大时，还可能更换目录存储并更新 dirPtr、dirLen 和目录深度。

表自己的 localDepth 可以小于 Map.globalDepth，这就是多个目录项能够共同指向同一个 table 的原因。增长处理必须保持这些字段的一致性，不能独立修改指针或长度。

## 6. 赋值与并发

map 赋值不会为每个副本深拷贝目录和条目，副本通常关联同一个 Map 控制结构。对条目的修改因此可能被其他副本观察到。

dirPtr 及目录更新不是供业务使用的同步机制。普通 map 并发读写仍需要正确同步，不能依赖内部指针更新或 writing 检测标记来保证安全。

## 7. 源码入口

- `src/internal/runtime/maps/map.go`：dirPtr、dirLen、目录索引、小 map 与目录增长相关逻辑。
- `src/internal/runtime/maps/table.go`：table 拆分及目录关系。
- `src/internal/runtime/maps/group.go`：单组的实际布局与访问。

上述说明对应 Go 1.27.2，后续版本可能调整内部字段与组织方式。
