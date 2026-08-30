# slice

## 1. 数据结构

Go runtime 在 `src/runtime/slice.go` 中定义了 slice 的内部结构：

```go
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}
```

slice 由三个字段组成：

- `array` 指向 slice 当前第一个元素。
- `len` 表示 slice 当前可以访问的元素数量。
- `cap` 表示从 `array` 指向的位置开始，到底层数组可用边界之间的元素数量。

## 2. 创建方式

### 2.1 零值声明

```go
var s []int
```

零值声明创建的是 nil slice：

```text
array = nil
len   = 0
cap   = 0
```

### 2.2 Slice 字面量

```go
s := []int{10, 20, 30}
```

编译器创建底层数组，并生成引用该数组的 slice。此时 `len` 和 `cap` 都为 `3`。

空字面量也属于这种创建方式：

```go
s := []int{}
```

它创建的是空 slice，但不是 nil slice。

### 2.3 使用 make 创建

```go
a := make([]int, 3)
b := make([]int, 3, 5)
```

创建结果为：

```text
a: len = 3, cap = 3
b: len = 3, cap = 5
```

`make` 是 slice 最主要的动态创建方式，底层可能涉及 `runtime.makeslice`。

### 2.4 从数组创建

```go
array := [5]int{10, 20, 30, 40, 50}
s := array[1:4]
```

该操作不会复制数组，`s` 引用 `array` 的一部分：

```text
array = &array[1]
len   = 3
cap   = 4
```

也可以从数组指针创建：

```go
array := &[5]int{10, 20, 30, 40, 50}
s := array[1:4]
```

### 2.5 从已有 Slice 创建

```go
source := []int{10, 20, 30, 40, 50}
s := source[1:4]
```

`source` 和 `s` 共享底层数组。

也可以使用完整切片表达式，通过第三个索引限制新 slice 的容量：

```go
s := source[1:4:4]
```

### 2.6 底层数组的栈分配条件

slice 不逃逸只表示底层数组具备栈分配资格，编译器还会检查容量是否可确定、所需空间是否过大，以及对应优化是否可用。

对于编译期容量为常量的 `make`，当前默认条件为：

```text
不逃逸
cap × elementSize <= 64 KiB
```

满足条件时，编译器可以将其改写为栈上数组：

```text
make([]T, len, cap) → var backing [cap]T; backing[:len]
```

对于容量在运行时才能确定的 `make` 或 `append`，当前编译器可以预留默认为 `32` 字节的栈缓冲区：

```text
K = 32 / elementSize

elementSize > 0
elementSize <= 32
make: runtimeCap <= K
append: newLen <= K
编译优化未被 -N 关闭
```

需要实际分配且不满足栈分配条件时，底层数组通常转为堆分配。slice 描述符需要单独判断；它可能保存在 SSA 值或寄存器中、spill 到栈上，或被编译器完全消除。

## 3. 扩容机制

### 3.1 扩容触发条件

slice 执行 `append` 时，追加后的长度超过当前容量才会触发扩容：

```text
newLen = oldLen + appendCount

newLen > oldCap
```

### 3.2 扩容流程

容量不足时，编译器调用 `runtime.growslice` 执行扩容：

1. 根据新长度和旧容量计算目标容量。
2. 根据元素大小计算所需字节数，并按内存分配规格调整实际容量。
3. 分配新的底层数组。
4. 将旧元素复制到新的底层数组。
5. 返回包含新数组指针、新长度和新容量的 slice，再写入本次追加的元素。

### 3.3 容量计算

标准扩容路径通过 `runtime.nextslicecap` 计算候选容量：

```text
如果 newLen > 2 × oldCap：
    candidateCap = newLen

否则，如果 oldCap < 256：
    candidateCap = 2 × oldCap

否则：
    candidateCap += (candidateCap + 3 × 256) >> 2
    重复计算，直到 candidateCap >= newLen
```

当容量从 `256` 开始增长时，该公式的增长倍数从 `2` 平滑过渡，并随容量增大逐渐趋近 `1.25`。

候选容量还需要转换为字节数，并通过 `roundupsize` 向上取整到内存分配器支持的规格：

```text
requestedBytes = candidateCap × elementSize
allocatedBytes = roundupsize(requestedBytes)
actualCap      = allocatedBytes / elementSize
```

因此，最终容量可能大于 `nextslicecap` 计算的候选容量，并受元素大小和内存分配规格影响。
