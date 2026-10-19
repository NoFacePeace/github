# Map globalShift

globalShift 是 Map 中的 uint8 字段，保存目录查找时需要对哈希值右移的位数。它由哈希位宽和 globalDepth 推导，用于提取目录索引。本文按 Go 1.27.2 官方源码说明。

## 1. 与 globalDepth 的关系

```text
globalShift = 哈希位宽 - globalDepth
```

globalDepth 表示取最高多少位，globalShift 表示右移多少位才能留下这些高位。常见 64 位平台使用 64 位哈希，32 位平台及 Wasm 使用 32 位哈希。

例如 64 位哈希、globalDepth 为 3 时，globalShift 为 61：

```text
哈希值：[101][剩余 61 位]
    ↓ 右移 61 位
目录索引：101（二进制），即 5
```

## 2. 目录查找

常规目录查找使用缓存的移位量提取索引：

```go
if m.dirLen == 1 {
    return 0
}
return hash >> (m.globalShift & 63)
```

常规模式 globalDepth 为零时只有一个目录项，源码通过 `dirLen == 1` 专门返回索引零。移位表达式中的 `& 63` 限定移位量范围。小 map 单组模式不使用这套目录选择 table 的路径。

## 3. 为什么单独保存

globalShift 是可以计算的派生值，单独保存相当于缓存移位量。目录深度相对较少变化，而目录查找是频繁执行的路径；缓存后，查找代码不必每次重新计算位宽与深度之差。

这是源码体现的热点路径简化方式，不能只根据少一次减法就断言存在很大的性能收益。实际效果还取决于编译优化、目标架构和负载。

## 4. 更新与一致性

目录初始化或扩大时，运行时根据 globalDepth 重新计算 globalShift，并与目录指针、长度等字段按协议更新。

这些字段不能由业务代码独立修改。若移位量与目录深度不一致，可能得到错误的索引，破坏查找或内存访问。

## 5. 存储成本

字段本身是 uint8，保存一个字节的值，但不能据此断言移除它一定让 Map 结构减少一个字节。结构体大小还受字段对齐和尾部填充影响。

## 6. 源码入口

- `src/internal/runtime/maps/map.go`：globalShift、depthToShift、directoryIndex 及目录增长时的更新。
