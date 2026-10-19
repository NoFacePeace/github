# Map 控制结构

`internal/runtime/maps.Map` 是 Go map 的顶层控制结构，维护元素总数、哈希种子、表目录及操作状态。本文按 Go 1.27.2 官方源码说明，它与语言层面 `map[K]V` 的使用语法不是同一层概念。

## 1. 结构定义

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

Map 本身不直接包含全部键值槽位，通常通过目录关联 table，再由表关联组数组；小 map 可以直接关联单组。

单组的控制信息与键值布局参见[Map group](group.md)。

## 2. 元素数量与哈希种子

| 字段 | 作用 |
| --- | --- |
| used | 全 map 有效条目数量，不包含删除槽 |
| seed | 每个 map 的随机哈希种子 |

used 位于结构首部，编译器与运行时按约定使用它支持 len 等操作。nil map 的 len 为零，但不能据此推断 nil map 已分配一个 used 为零的控制结构。

哈希计算结合键、类型对应的哈希算法和 seed。两个 map 保存相同键，也可能因种子不同而将它们安排到不同存储位置；种子不是遍历顺序的稳定保证。

## 3. 目录相关字段

| 字段 | 作用 |
| --- | --- |
| [dirPtr](dirPtr.md) | 通常指向 table 指针目录，小 map 时可直接指向 group |
| [dirLen](dirLen.md) | 目录项数量；单组模式时为零 |
| [globalDepth](globalDepth.md) | 用于目录选择的哈希前缀位数 |
| [globalShift](globalShift.md) | 提取目录索引时使用的移位量 |

常规模式下目录长度为 `1 << globalDepth`，多个目录项可以指向同一个 table。字段的解释需要结合模式，不能只看 dirPtr 就确定其目标类型。

## 4. 写入检测标记

writing 在写入过程中按内部协议翻转，用于提高发现并发误用的机会。它不是 Mutex，不会把同时到来的写者正常排队，也不保证检测到所有竞争。

普通 map 的并发读写仍需外部同步；不能因为结构含有 writing 就认为 map 天然支持并发更新。

## 5. 删除标记与清空序列

tombstonePossible 为 false 表示已知整个 map 不含删除标记；为 true 则表示可能存在，不等于精确的删除槽数量。实际删除槽状态仍由组内控制信息记录。

clearSeq 记录 clear 操作的序列，供迭代等路径识别清空事件。它不是有效条目计数，也不是可供业务依赖的版本号或并发同步工具。

## 6. 与 map 值的关系

赋值和传参复制 map 值，不深拷贝控制结构和全部条目，因此副本可以继续关联同一个映射。向一个副本写入条目，可以被另一个副本观察到；将某个局部变量重新赋给新 map，则不会自动替换其他副本。

已初始化的空 map 可以尚未分配键值组，不能把“没有组存储”与“语言层面是 nil map”画等号。实际分配位置和时机还受编译优化影响。

## 7. 源码入口

- `src/internal/runtime/maps/map.go`：Map 定义、目录、小 map、清空等操作。
- `src/cmd/compile/internal/reflectdata/map.go`：编译器侧相关类型布局约定。

这些字段属于 Go 1.27.2 的内部实现，不能通过 unsafe 修改并期待跨版本兼容。
