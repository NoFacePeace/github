# map

当前 map 的底层实现位于 `src/internal/runtime/maps`，采用 Swiss Table，而不是旧版的 bucket 和 overflow bucket。

## 1. 数据结构

map 变量保存的是指向 runtime `Map` 的指针。赋值和传参只会复制这个指针，因此多个 map 变量可以访问同一份数据；map 不能比较，只有 map 与 `nil` 可以比较。

`src/internal/runtime/maps/map.go` 中的核心结构可以简化为：

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
	clearSeq           uint64
}
```

- `used` 是元素数量，也是 `len(m)` 的数据来源。
- `seed` 是每个 map 独立的随机哈希种子。
- `dirPtr` 和 `dirLen` 描述 table 目录；小 map 使用 `dirLen == 0` 表示特殊布局。
- `globalDepth` 和 `globalShift` 用于从哈希高位选择 table。
- `writing` 用于尽可能检测并发读写或并发写入，不是互斥锁。
- `tombstonePossible` 表示 table 中是否可能存在删除标记。
- `clearSeq` 帮助迭代器识别迭代期间发生的 `clear`。

### 1.1 整体层次

完整 map 的层次为：

```text
map value
    │
    ▼
  *Map ──→ directory: []*table
                         │
                         ▼
                       table ──→ groupsReference
                                      │
                                      ▼
                                    group
                                      ├── 8-byte control word
                                      └── 8 × slot(key, elem)
```

`Map` 管理全局状态并选择 table，table 是一张能够独立查找和扩容的 Swiss Table，group 是一次批量匹配的基本单位，slot 才真正保存 key 和 elem。

### 1.2 Map

`Map` 是顶层描述符，不直接保存普通 map 的 key 和 elem。它主要负责：

- 保存总元素数和随机哈希种子。
- 持有 directory 或小 map 的单个 group。
- 根据哈希高位选择 table。
- 协调 table 分裂、directory 扩容和迭代状态。

`used` 必须是第一个字段，因为编译器生成的 `len(m)` 会直接读取它。`seed` 使相同类型、相同内容的两个 map 也拥有不同哈希分布，既能减少构造碰撞攻击的风险，也会影响遍历布局。

### 1.3 小 map

元素数量不超过 8 时可以使用特殊布局：

```text
Map.dirPtr ──→ group
Map.dirLen = 0
```

此时没有 directory 和 table，直接在一个 group 中线性处理候选 slot。小 map 没有跨 group 的探测序列，因此删除可以直接将 slot 恢复为 empty，不需要 tombstone。

第 9 个元素写入时，小 map 转换为容量为 16 个 slot 的完整 table。

### 1.4 Directory

directory 不是单独的 Go struct，而是一段 `[]*table` 指针数组：

```text
dirPtr ──→ [*table, *table, *table, ...]
dirLen  = 1 << globalDepth
```

runtime 取哈希最高 `globalDepth` 位作为目录索引。在 64 位平台上概念上等价于：

```text
directoryIndex = hash >> (64 - globalDepth)
```

多个连续目录项可以指向同一个 table。例如 directory 已扩容，但某个 table 尚未分裂时：

```text
00 ─┐
01 ─┴─→ table A, localDepth = 1
10 ───→ table B, localDepth = 2
11 ───→ table C, localDepth = 2
```

这种共享使 directory 可以独立扩容，不必同时分裂所有 table。

### 1.5 Table

`src/internal/runtime/maps/table.go` 中的 table 可以简化为：

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

- `used`：当前有效 slot 数，不包含 tombstone。
- `capacity`：slot 总数，始终是 2 的幂，最大为 1024。
- `growthLeft`：在扩容或重新散列前还能占用的 empty slot 数；tombstone 也会消耗增长空间。
- `localDepth`：该 table 已经使用了多少个哈希高位，用于判断分裂时是否还要扩大 directory。
- `index`：table 在 directory 中第一次出现的位置；被替换的旧 table 使用 `-1`。
- `groups`：连续 group 数组的描述符。

每个 table 都是一张完整的开放寻址哈希表，可以独立查找、扩容或分裂。`globalDepth` 属于整个 Map，`localDepth` 属于单个 table，两者分离是 extendible hashing 能够局部增长的关键。

### 1.6 GroupsReference 与 Group

`groupsReference` 描述一段连续的 group 数组：

```go
type groupsReference struct {
	data       unsafe.Pointer
	lengthMask uint64
}
```

group 数量始终是 2 的幂，因此 `lengthMask = groupCount - 1`，探测位置取模可以转化为更便宜的按位与：

```text
groupIndex = probeOffset & lengthMask
```

group 的具体类型由编译器根据 map 的 key 和 elem 类型生成。每个 group 固定包含 8 个 slot 和 8 个 control byte。它可以采用 key/elem 交错布局：

```text
ctrls | key0 elem0 | key1 elem1 | ... | key7 elem7
```

启用 `GOEXPERIMENT=mapsplitgroup` 时也可以采用分离布局：

```text
ctrls | key0 ... key7 | elem0 ... elem7
```

交错布局有利于命中 key 后立即读取 elem；分离布局在只扫描 key 或 key/elem 对齐差异较大时可能拥有更好的空间和缓存行为。runtime 通过 `MapType` 中的 offset 和 stride 统一访问两种布局。

### 1.7 Control Word

control word 是一个 `uint64`，由 8 个 control byte 组成，每个字节对应同索引的 slot：

```text
empty:   1000_0000
deleted: 1111_1110
full:    0hhh_hhhh
```

满 slot 的低 7 位保存哈希的 H2 部分。查找时，runtime 对整个 control word 做位运算，在不读取 key 的情况下同时得到 8 个 slot 的候选位图：

```text
matchH2(H2)           → H2 相同的 slot
matchEmpty()          → empty slot
matchEmptyOrDeleted() → 可以用于插入的 slot
```

AMD64 上部分匹配操作会替换为 SIMD intrinsic。control word 将热点元数据集中在 8 字节中，是 Swiss Table 减少缓存访问和分支的核心。

### 1.8 Slot

一个 slot 在逻辑上保存一对 key 和 elem：

```go
struct {
	key  K
	elem V
}
```

key 或 elem 大于 128 字节时，slot 改为保存指针，实际对象单独分配；较小对象通常直接内联。间接存储限制了 group 大小和搬迁成本，但增加了一次指针间接访问和独立分配。

slot 自身不保存完整哈希值，只有对应 control byte 保存 H2。最终确认匹配仍需调用 key 类型的相等函数。

### 1.9 MapType

每一种 `map[K]V` 类型都有编译器生成的 `internal/abi.MapType` 元数据，关键内容包括：

- key、elem 和 group 的类型信息。
- key 哈希函数 `Hasher` 和相等函数。
- group 大小以及 key/elem 的 offset、stride。
- key 或 elem 是否间接存储等标志。

`Map` 保存实例状态，`MapType` 保存同一种 map 类型共享的操作和布局信息。runtime 的入口通常同时接收两者：

```text
runtime.mapaccess2(mapType, mapInstance, key)
```

### 1.10 哈希如何连接各层

哈希值被拆成 H1 和 H2：

```text
hash
┌──────────────────────── H1 ───────────────────────┬── H2 ──┐
│             directory 与 group 探测               │ 低 7 位 │
└───────────────────────────────────────────────────┴─────────┘
```

- 哈希最高位根据 `globalDepth` 选择 directory 中的 table。
- H1 配合 `groups.lengthMask` 构造 table 内的 group 探测序列。
- H2 与 control word 批量匹配，产生需要执行完整 key 比较的 slot。

因此一次完整定位过程是：

```text
hash(key, seed) → directory → table → probe group → match H2 → compare key → elem
```

## 2. 创建与零值

### 2.1 nil map

```go
var m map[string]int
```

nil map 可以读取、查询长度、遍历和删除；读取返回元素类型零值，写入会 panic。

### 2.2 使用 make 创建

```go
m := make(map[string]int)
n := make(map[string]int, 100)
```

容量参数是元素数量提示，不是固定容量。编译器将创建过程降低为 `runtime.makemap_small` 或 `runtime.makemap`，runtime 根据 hint 预分配能够容纳这些元素的结构。

如果 map 不逃逸，编译器可以将 `Map` 放在栈上；当 hint 不超过 8 时，还可以同时在栈上预留第一个 group。map 后续增长所需的 table 和 group 仍可能在堆上分配。

### 2.3 字面量

```go
m := map[string]int{
	"alice": 1,
	"bob":   2,
}
```

字面量不使用另一套底层结构。编译器会先创建 map，再执行对应的元素写入。

#### 2.3.1 逃逸分析

编译器前端使用 `OMAPLIT` 表示 map 字面量。`src/cmd/compile/internal/escape/expr.go` 会直接对 `OMAPLIT` 执行逃逸分析，因此是否在源码中显式出现 `make` 不影响栈或堆的判断。

需要分别判断两类逃逸：

- 字面量产生的 `Map` 是否逃逸；不逃逸时，它具备栈分配资格。
- key 或 elem 引用的对象是否被 map 保存；即使 `Map` 在栈上，这些对象仍可能逃逸到堆上。

nil map 不创建 `Map`、table 或 group，因此没有底层结构的栈/堆选择。空字面量则会创建一个非 nil map：

```go
var nilMap map[string]int
emptyMap := map[string]int{}
```

```text
nilMap == nil   → true
emptyMap == nil → false
```

#### 2.3.2 降低为 make 和写入

`src/cmd/compile/internal/walk/complit.go:maplit` 将字面量概念上降低为：

```go
m := make(map[string]int, 2)
m["alice"] = 1
m["bob"] = 2
```

编译器生成内部 `OMAKEMAP` 节点时，会将 `OMAPLIT` 的逃逸结果复制给它：

```text
OMAPLIT
→ escape analysis
→ maplit 生成 OMAKEMAP
→ OMAKEMAP 继承 OMAPLIT 的逃逸结果
→ walkMakeMap
```

因此，不逃逸且条目数不超过 8 的字面量可以复用 `make(map)` 的栈优化：

```text
当前函数栈帧
├── Map
└── 首个 8-slot group
```

超过 8 个条目时，即使 `Map` 描述符不逃逸，当前编译器通常也只将 `Map` 放在栈上，directory、table 和 group 由 runtime 分配在堆上。

#### 2.3.3 静态与动态条目

对于包含运行时计算的字面量：

```go
m := map[int]int{
	1:       10,
	loadKey(): loadValue(),
	3:       30,
}
```

`src/cmd/compile/internal/walk/order.go` 将动态条目拆出，并按源码求值顺序逐项写入：

```go
m := map[int]int{
	1: 10,
	3: 30,
}
m[loadKey()] = loadValue()
```

动态条目数量仍会计入 `make` 的容量 hint，尽量避免字面量初始化过程中发生扩容。逐项处理还保证前一个 key/value 的求值和写入完成后，才开始计算后一个条目。

#### 2.3.4 大字面量初始化

完成动态条目拆分后，如果剩余静态条目不超过 25 个，编译器直接生成逐项赋值：

```text
tmpKey = key0; tmpElem = elem0; m[tmpKey] = tmpElem
tmpKey = key1; tmpElem = elem1; m[tmpKey] = tmpElem
...
```

如果静态条目超过 25 个，逐项展开会显著增大机器码。编译器改为生成两份只读静态数组，并通过循环初始化 map：

```go
staticKeys  := [...]K{key0, key1, key2}
staticElems := [...]V{elem0, elem1, elem2}

for i := 0; i < len(staticKeys); i++ {
	m[staticKeys[i]] = staticElems[i]
}
```

这种优化减少代码体积，但每个元素仍经过正常的 map 写入路径和哈希计算；它不是在编译期直接构造 Swiss Table，因为实际布局依赖每个 map 运行时生成的随机 `seed`。

## 3. 查找机制

`value, ok := m[key]` 的主要调用链为：

```text
编译器降低
→ runtime.mapaccess2 或特化的 fast32/fast64/faststr 入口
→ internal/runtime/maps
→ table/group 探测
```

完整 map 的查找流程：

1. 使用 key、key 类型对应的哈希函数和 `seed` 计算哈希值。
2. 使用哈希高位在 directory 中选择 table。
3. 使用 H1 构造以 group 为单位的二次探测序列。
4. 将 H2 与 group 的 8 个 control byte 并行匹配，只对候选 slot 执行 key 相等比较。
5. 找到相等 key 时返回 elem；遇到含 empty slot 的 group 时结束查找。

探测序列按三角数递增：

```text
p(i) = H1 + (i² + i) / 2 mod groupCount
```

group 数量始终是 2 的幂，因此该序列能够遍历 table 中的所有 group。control word 先过滤绝大多数不匹配 slot，可以减少分支和昂贵的 key 相等比较。

元素地址可能在写入、删除或扩容后变化，所以 Go 不允许直接取得 `m[key]` 的地址。

## 4. 写入与删除

写入由 `runtime.mapassign` 等入口完成：先按查找流程定位 key；存在时返回原 elem slot，不存在时优先使用探测路径中的 deleted slot，否则使用第一个 empty slot。

删除时必须保持开放寻址的探测不变量：

- 如果 group 仍有 empty slot，可以将目标 control byte 直接改为 empty。
- 如果 group 没有 empty slot，直接改为 empty 可能让后续查找提前停止，因此需要标记为 deleted tombstone。

tombstone 会增加探测成本。写入空间不足时，runtime 会尝试批量清理不再需要的 tombstone；收益不足时执行扩容或 table 分裂。

## 5. 扩容机制

普通 table 的最大平均负载因子是 `7/8`。每个 table 独立计算剩余增长空间，并独立扩容：

1. 小 map 最多直接存放在单个 8-slot group 中；放不下时转换为 16-slot table。
2. table 容量未超过 1024 个 slot 时，分配两倍容量的新 table，并重新散列已有元素。
3. table 达到最大容量后，将其按一个新的哈希高位分裂为两个 table。
4. table 分裂需要更多目录位时，directory 扩大一倍；未分裂的 table 可以被多个连续目录项引用。

这种 extendible hashing 只增长命中的 table，避免每次都重建整个 map，降低大 map 单次扩容的延迟；代价是多一层 directory 间接访问，以及更复杂的迭代和增长状态维护。

## 6. 遍历与并发

map 遍历顺序未定义，runtime 会随机化起始位置。迭代期间发生 table 扩容时，迭代器继续扫描旧 table 来避免重复返回 key，同时到新 table 中重新查询 key，以获得最新值并跳过已删除项。

普通 map 不提供并发同步：

- 多个 goroutine 只读同一个不再修改的 map 是安全的。
- 读写并发或多个写操作并发必须由调用方同步。
- `writing` 标志只能帮助 runtime 检测部分非法并发并终止程序，不能替代锁，也不能保证检测所有数据竞争。

需要并发读写时，应根据访问模式使用 `sync.RWMutex` 保护普通 map，或使用适合其特定场景的 `sync.Map`。

## 7. 源码位置

- `src/internal/runtime/maps/map.go`：`Map`、小 map、目录和顶层操作。
- `src/internal/runtime/maps/table.go`：查找、插入、删除、扩容、分裂和迭代。
- `src/internal/runtime/maps/group.go`：group、control word 和批量匹配。
- `src/internal/runtime/maps/runtime.go`：通用 runtime map 操作入口。
- `src/internal/runtime/maps/runtime_fast*.go`：整数、指针和字符串 key 的特化路径。
- `src/runtime/map.go`：编译器可见的 runtime 入口和创建逻辑。
- `src/cmd/compile/internal/walk/builtin.go`：`make(map)` 的编译期栈分配优化。
- `src/cmd/compile/internal/escape/expr.go`：map 创建与字面量的逃逸分析。
- `src/cmd/compile/internal/walk/order.go`：字面量动态条目的求值和写入顺序。
- `src/cmd/compile/internal/walk/complit.go`：字面量降低和大字面量循环初始化。
