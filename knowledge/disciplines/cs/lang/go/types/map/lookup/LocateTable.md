# Map 定位 table

常规目录模式下，Go map 使用键的哈希前缀选择目录项，再读取该目录项保存的 table 指针。此步骤定位目标表，不负责确定键是否存在。本文按 Go 1.27.2 官方源码说明。

## 1. 前提与输入

运行时先处理 nil、空 map 等情况，并结合键、类型对应的哈希算法和 Map.seed 计算哈希值。

`dirLen == 0` 时处于小 map 单组模式，直接检查 group，不执行常规目录查找；`dirLen > 0` 时才根据目录选择 table。

## 2. 提取哈希高位

目录使用最高 globalDepth 位，globalShift 预先保存移位量：

```text
globalShift = 哈希位宽 - globalDepth
```

常见 64 位平台使用 64 位哈希；32 位平台及 Wasm 使用 32 位哈希。64 位哈希、globalDepth 为 3 时：

```text
哈希：[101][剩余 61 位]
    ↓ 右移 61 位
目录索引：5
```

相关字段分别参见[globalDepth](../structs/globalDepth.md)、[globalShift](../structs/globalShift.md)和[dirLen](../structs/dirLen.md)。

## 3. directoryIndex

Go 1.27.2 的源码为：

```go
func (m *Map) directoryIndex(hash uintptr) uintptr {
    if m.dirLen == 1 {
        return 0
    }
    return hash >> (m.globalShift & 63)
}
```

只有一个目录项时直接返回零。其他情况下通过右移保留所需的高位，`& 63` 限定移位量范围。这个方法不是用哈希对 table 数量取模，也不直接返回 table 指针。

## 4. directoryAt

拿到索引后，通过 dirPtr 读取目录项：

```go
func (m *Map) directoryAt(i uintptr) *table {
    return *(**table)(unsafe.Pointer(
        uintptr(m.dirPtr) + goarch.PtrSize*i,
    ))
}
```

目录是 table 指针数组，每项占一个指针大小。代码计算第 i 项的地址，再读取其中保存的 table 指针。索引合法性依赖目录深度、长度及调用协议，不能把此内部指针运算当作业务侧通用访问方式。

目录指针的两种含义参见[dirPtr](../structs/dirPtr.md)，目标表的结构参见[Map table](../structs/table.md)。

## 5. 多个目录项可以指向同一张表

```text
目录索引： 0   1   2   3
目录内容：[A][A][B][B]
```

不同哈希前缀可以选中同一 table，具体对应关系由 table.localDepth 与 Map.globalDepth 共同决定。目录有 4 项不代表有 4 张独立的表。

选定 table 后，查找仍需进行组探测、H2 指纹匹配和完整键比较。定位成功不等于查找命中。

后续起始组索引与地址计算参见[Map 定位 group](LocateGroup.md)。

## 6. 源码入口

- `src/internal/runtime/maps/map.go`：哈希计算、directoryIndex 和 directoryAt。
- `src/internal/runtime/maps/table.go`：定位 table 后的表内查找。
