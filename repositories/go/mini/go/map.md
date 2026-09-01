# map

当前 map 的底层实现位于 `src/internal/runtime/maps`，采用 Swiss Table，并在其上使用 extendible hashing 管理多个 table。

## 1. 语言语义

### 1.1 nil map

```go
var m map[string]int
```

nil map 没有底层 `Map`、table 或 group：

```text
m = nil
```

它可以读取、查询长度、遍历、删除和执行 `clear`；读取返回元素类型零值，写入会 panic。

### 1.2 非 nil map

`make` 和字面量都会创建非 nil map：

```go
emptyMap := make(map[string]int)
literalMap := map[string]int{}
```

容量 hint 只是预计元素数量，不是固定容量：

```go
m := make(map[string]int, 100)
```

### 1.3 引用语义

map 变量保存的是指向 runtime `Map` 的指针。赋值和传参只复制这个指针，因此多个变量可以访问同一份数据：

```go
source := map[string]int{"count": 1}
target := source
target["count"] = 2
```

map 不能相互比较，只能与 `nil` 比较；`len(m)` 返回当前元素数量，遍历顺序未定义。

## 2. 数据结构骨架

### 2.1 整体层次

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

`Map` 管理全局状态并选择 table；table 是一张能够独立查找和增长的 Swiss Table；group 是批量匹配的基本单位；slot 保存 key 和 elem。

### 2.2 Map

`src/internal/runtime/maps/map.go` 中的顶层结构可以简化为：

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

- `used`：有效元素总数，也是 `len(m)` 的数据来源。
- `seed`：每个 map 独立的随机哈希种子。
- `dirPtr`、`dirLen`：directory 或小 map group 的入口。
- `globalDepth`、`globalShift`：从哈希高位计算 directory 索引。
- `writing`：帮助检测非法并发写入，不是互斥锁。
- `tombstonePossible`：table 中是否可能存在删除标记。
- `clearSeq`：帮助迭代器识别迭代期间发生的 `clear`。

`used` 必须位于第一个字段，编译器生成的 `len(m)` 会直接读取它。随机 `seed` 使相同类型和内容的不同 map 也具有不同哈希分布。

### 2.3 小 map

最多 8 个元素可以直接存放在单个 group 中：

```text
Map.dirPtr ──→ group
Map.dirLen = 0
```

此时没有 directory 和 table，也没有跨 group 探测，所以删除不需要 tombstone。第 9 个元素写入时，小 map 转换为完整 table。

### 2.4 Directory

directory 不是独立 struct，而是 `[]*table` 指针数组：

```text
dirPtr ──→ [*table, *table, *table, ...]
dirLen  = 1 << globalDepth
```

runtime 取哈希最高 `globalDepth` 位作为目录索引。在 64 位平台上概念上等价于：

```text
directoryIndex = hash >> (64 - globalDepth)
```

多个连续目录项可以指向同一个 table：

```text
00 ─┐
01 ─┴─→ table A, localDepth = 1
10 ───→ table B, localDepth = 2
11 ───→ table C, localDepth = 2
```

这使 directory 能够独立扩大，而不必同时分裂所有 table。

### 2.5 Table

`src/internal/runtime/maps/table.go` 中的结构可以简化为：

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

- `used`：有效 slot 数，不包含 tombstone。
- `capacity`：slot 总数，始终是 2 的幂，最大为 1024。
- `growthLeft`：重新散列前还能占用的 empty slot 数，tombstone 也消耗该额度。
- `localDepth`：当前 table 已使用的哈希高位数量。
- `index`：table 在 directory 中第一次出现的位置。
- `groups`：连续 group 数组的描述符。

每个 table 都能独立查找、扩容或分裂。Map 的 `globalDepth` 与 table 的 `localDepth` 分离，是 extendible hashing 能够局部增长的关键。

### 2.6 Group 与 Control Word

`groupsReference` 描述连续 group 数组：

```go
type groupsReference struct {
	data       unsafe.Pointer
	lengthMask uint64
}
```

group 数量始终是 2 的幂，所以 `lengthMask = groupCount - 1`，取模可以转换为按位与：

```text
groupIndex = probeOffset & lengthMask
```

每个 group 包含 8 个 slot 和一个由 8 个 control byte 组成的 `uint64`：

```text
empty:   1000_0000
deleted: 1111_1110
full:    0hhh_hhhh
```

`full` 表示对应 slot 保存有效 key/elem，不代表整个 group 已满；只有 8 个 slot 全部为 full，group 才是满的。满 slot 的低 7 位保存 H2。

runtime 可以对整个 control word 执行批量匹配：

```text
matchH2(H2)           → H2 相同的 slot
matchEmpty()          → empty slot
matchEmptyOrDeleted() → 可用于插入的 slot
```

AMD64 上部分匹配操作会替换为 SIMD intrinsic。热点元数据集中在 8 字节中，可以减少 key 读取、分支和缓存访问。

### 2.7 Slot 与 Group 布局

slot 在逻辑上保存一对 key 和 elem。key 或 elem 大于 128 字节时，slot 保存指针，实际对象单独分配；较小对象通常内联。

关闭 `GOEXPERIMENT=mapsplitgroup` 时采用 key/elem 交错布局：

```text
ctrls | key0 elem0 | key1 elem1 | ... | key7 elem7
```

当前默认启用 `GOEXPERIMENT=mapsplitgroup`，采用分离布局：

```text
ctrls | key0 ... key7 | elem0 ... elem7
```

runtime 通过 `MapType` 中的 offset 和 stride 统一访问两种布局。slot 不保存完整哈希，最终确认匹配仍需调用 key 类型的相等函数。

### 2.8 MapType 与哈希分工

每一种 `map[K]V` 类型都有编译器生成的 `internal/abi.MapType`，其中保存：

- key、elem 和 group 的类型信息。
- key 哈希函数 `Hasher` 和相等函数。
- group 大小以及 key/elem 的 offset、stride。
- key 或 elem 是否间接存储等标志。

哈希被拆成 H1 和 H2：

```text
hash
┌──────────────────────── H1 ───────────────────────┬── H2 ──┐
│             directory 与 group 探测               │ 低 7 位 │
└───────────────────────────────────────────────────┴─────────┘
```

```text
hash(key, seed) → directory → table → probe group → match H2 → compare key → elem
```

## 3. 创建与初始化

### 3.1 状态变化

map 创建和首次写入可能经历：

```text
nil map
   │ make 或字面量
   ▼
empty Map
   │ 首次写入
   ▼
small map: 单个 8-slot group
   │ 第 9 个元素
   ▼
full map: directory → table → groups
```

nil map 不会在写入时自动创建，因此第一步必须由 `make` 或字面量完成。

### 3.2 创建空 map

```go
m := make(map[string]int)
```

runtime 路径主要是：

```text
runtime.makemap_small
→ maps.NewEmptyMap
→ 创建 Map 并初始化随机 seed
```

堆分配路径不会立即创建 group。第一次写入发现 `dirPtr == nil` 时调用 `growToSmall` 分配首个 group。编译器确认 map 不逃逸时，可以提前在栈上预留 `Map` 和首个 group，具体优化放在第 9 节。

### 3.3 使用容量 hint

```go
m := make(map[string]int, hint)
```

主要调用链为：

```text
runtime.makemap
→ maps.NewMap
```

- `hint <= 8`：保持小 map 资格，不需要 directory 和 table。
- `hint > 8`：按 `7/8` 负载因子估算容量，创建 directory、table 和 group 数组。

hint 只用于减少初始化阶段的扩容，不能限制最终元素数量。

### 3.4 字面量初始化

```go
m := map[string]int{
	"alice": 1,
	"bob":   2,
}
```

字面量使用相同底层结构，概念上分为：

```text
make(map[string]int, 2)
→ 写入 "alice"
→ 写入 "bob"
```

它不是在编译期直接生成最终 Swiss Table，因为实际布局依赖运行时生成的随机 `seed`。详细降低过程放在第 9 节。

## 4. 查找

`value, ok := m[key]` 的主要调用链为：

```text
编译器降低
→ runtime.mapaccess2 或 fast32/fast64/faststr 入口
→ internal/runtime/maps
→ table/group 探测
```

完整 map 的查找流程：

1. 使用 key、类型哈希函数和 `seed` 计算 hash。
2. 使用 hash 高位在 directory 中选择 table。
3. 使用 H1 构造以 group 为单位的探测序列。
4. 将 H2 与 group 的 8 个 control byte 批量匹配。
5. 只对候选 full slot 执行完整 key 比较。
6. 找到相同 key 时返回 elem；遇到 empty 时结束查找。

### 4.1 三角数探测

探测步长依次为 `+1、+2、+3`，累计偏移是三角数：

```text
0, 1, 3, 6, 10, 15, ...
```

```text
p(i) = H1 + i × (i + 1) / 2 mod groupCount
```

group 数量为 2 的幂时，该序列能够访问所有 group。它保留一定缓存局部性，同时降低线性探测在连续写入冲突时形成主聚集的程度。

### 4.2 终止条件

- `full`：可能是目标 key，需要比较 H2 和完整 key。
- `deleted`：目标 key 可能在后续 group，不能停止。
- `empty`：插入时总会使用此前第一个可用位置，因此目标 key 不可能越过 empty，查找可以停止。

元素地址可能在写入、删除或扩容后变化，所以 Go 不允许直接取得 `m[key]` 的地址。

## 5. 写入与冲突

写入由 `runtime.mapassign` 等入口完成，核心是在执行查找的同时记录可用位置。

### 5.1 相同 key 与哈希冲突

- H2 和完整 key 都相等：返回原 elem slot，覆盖 elem，元素数量不变。
- H2 相等但 key 不相等：这是 H2 假阳性，继续检查其他候选 slot。
- 初始 group 已占用：按照三角数序列继续探测其他 group。

### 5.2 Empty 与 Deleted

写入在一个 group 中先比较所有 H2 候选，再检查 empty 或 deleted：

```text
找到相同 key → 更新 elem
找到 deleted  → 记住第一个位置，继续查找
找到 empty    → 确认 key 不存在
               → 使用之前的 deleted，否则使用当前 empty
```

不能遇到 deleted 就立即插入，因为相同 key 可能位于后续探测路径，否则可能产生两个相等 key。复用 deleted 还能缩短探测路径，并且不额外消耗 `growthLeft`。

### 5.3 空间不足

使用新的 empty slot 需要 `growthLeft > 0`。额度耗尽时，runtime 先尝试清理 tombstone；不能回收足够空间时重新散列、扩容或分裂 table，然后重新定位插入位置。

## 6. 删除

### 6.1 删除后的状态

删除先按正常查找路径定位 key，然后清理 key 和 elem。control byte 的处理必须维持开放寻址的不变量：

- group 已经存在 empty：将目标 slot 改为 empty，并增加 `growthLeft`。
- group 没有 empty：将目标 slot 改为 deleted，避免后续查找被提前终止。

小 map 没有跨 group 探测，删除可以直接恢复 empty，不需要 tombstone。

tombstone 会延长查找路径并消耗增长额度，但不能直接全部改成 empty，否则可能截断仍然有效的探测链。

## 7. 扩容

当前实现的 tombstone 清理、table 容量翻倍、table 分裂和 directory 扩大，都由触发操作的 goroutine 在本次写入内同步完成。它不会把迁移工作分摊给后续写入，也不是全局 Stop-The-World，但会增加本次写入延迟。

### 7.1 触发条件

更新已有 key 不增加元素数量，不会触发扩容。只有写入不存在的新 key 时，才需要检查下面两种情况。

#### 7.1.1 小 map

小 map 已有 8 个元素时，第 9 个不同 key 无处插入，必须转换为 1 个容量为 16 个 slot、包含 2 个 group 的完整 table。

#### 7.1.2 普通 table

普通 table 写入新 key 时，如果需要使用新的 empty slot，并且 `growthLeft == 0`，runtime 会先尝试清理 tombstone；清理后仍没有增长额度才真正扩容。

容量大于 8 的普通 table 使用：

```text
maxGrowth = capacity × 7 / 8
growthLeft = maxGrowth - used - tombstones
```

因此触发条件不是 `used == capacity`，而是有效元素和 tombstone 已经共同耗尽 `7/8` 负载额度。探测过程中如果能够复用 deleted slot，就不消耗新的 empty，也不会触发扩容。

```text
写入不存在的新 key
        │
        ├── 可以复用 deleted ───────────────→ 直接插入
        │
        ├── growthLeft > 0 ─────────────────→ 使用 empty
        │
        └── growthLeft == 0
                │
                ├── 清理 tombstone 后恢复额度 → 使用 empty
                └── 仍无额度 ─────────────────→ 扩容
```

### 7.2 Tombstone 清理

写入新 key 需要使用 empty，但 `growthLeft == 0` 时，runtime 调用 `pruneTombstones` 尝试恢复增长额度。

清理过程为：

1. 计算 tombstone 数量；少于 table 容量的 10% 时直接放弃，避免为少量回收执行全表扫描。
2. 遍历所有 full slot，重新计算 key 的 hash，并重放从初始 group 到当前 group 的探测路径。
3. 探测路径必须经过的 tombstone group 标记为 needed，不能清理，否则查找会在新产生的 empty 提前终止。
4. 统计未标记 group 中可以安全清理的 tombstone；可回收数量少于容量的 10% 时放弃。
5. 将可清理的 deleted control byte 改为 empty；每清理一个，`growthLeft` 增加一。

```text
full entry 的有效探测路径
起点 → needed tombstone → needed tombstone → entry 所在 group
             │
             └── 必须保留，不能变成 empty

不被任何有效探测路径依赖的 tombstone
deleted → empty → growthLeft + 1
```

这个过程只修改 control byte，不移动有效元素，避免破坏迭代器正在使用的 slot 顺序。清理后 `growthLeft > 0` 就继续插入；仍为 0 则转入 table 扩容或分裂。

### 7.3 扩容流程

#### 7.3.1 小 map 转换

小 map 写入第 9 个不同 key 时，创建 1 个容量为 16 个 slot 的 table，其中包含 2 个 group。runtime 将原 group 中的 8 个元素重新散列到新 table，再插入新 key。

```text
small map: 1 × group
→ directory: 1 × table
→ table: 2 × group = 16 slots
```

#### 7.3.2 Table 容量翻倍

如果 table 容量小于 1024 个 slot，runtime 创建容量为原来两倍的新 table，重新散列全部有效元素，再替换 directory 中指向旧 table 的引用：

```text
16 → 32 → 64 → 128 → 256 → 512 → 1024
```

旧 table 被标记为失效，但迭代器可能继续持有它，以维持迭代期间不重复返回 key 的语义。

#### 7.3.3 Table 分裂

容量为 1024 的 table 再次需要增长时，不再扩大单个 table，而是将 `localDepth` 加一，并创建两个容量为 1024 的 table。已有元素按照新增的哈希高位重新分配到左右 table：

```text
旧 table 匹配前缀 P
→ hash 前缀 P0：放入 left table
→ hash 前缀 P1：放入 right table
```

如果 directory 已经有足够的索引位，旧 table 对应的连续目录区间直接分成两半。假设 `globalDepth = 3`，旧 table 的 `localDepth = 1`，匹配前缀 `0`：

```text
分裂前：
000 ─┐
001 ─┤
010 ─┤→ old table
011 ─┘

分裂后 localDepth = 2：
000 ─┐
001 ─┴→ left table，匹配前缀 00

010 ─┐
011 ─┴→ right table，匹配前缀 01
```

也就是旧区间的前半指向 left，后半指向 right。分裂只处理发生增长的局部 table，不需要重建其他 table。

#### 7.3.4 Directory 扩大

table 分裂时，如果：

```text
table.localDepth == map.globalDepth
```

说明 directory 没有多余索引位区分两个新 table。runtime 先将 directory 长度翻倍并增加 `globalDepth`，再安装分裂后的左右 table。

directory 翻倍时，每个旧索引复制成两个新索引，并暂时继续指向同一个 table：

```text
newDirectory[2 × i]     = oldDirectory[i]
newDirectory[2 × i + 1] = oldDirectory[i]
```

例如 `globalDepth` 从 2 增加到 3：

```text
旧索引 00 → 新索引 000、001 → 原 table A
旧索引 01 → 新索引 010、011 → 原 table B
旧索引 10 → 新索引 100、101 → 待分裂 old table
旧索引 11 → 新索引 110、111 → 原 table C
```

随后只替换待分裂 table 对应的两个新索引：

```text
100 → left table
101 → right table
```

其他 table 仍由复制后的多个连续目录项引用。如果 `localDepth < globalDepth`，directory 本来就有足够索引位，只需按上一节将旧区间分成左右两半，不需要扩大 directory。

```text
单个 table 空间不足
        │
        ├── capacity < 1024 → grow: 容量翻倍
        │
        └── capacity = 1024 → split: 分裂为两个 table
                                  │
                                  └── 必要时扩大 directory
```

这种设计只增长命中的 table，避免每次重建整个大 map，降低单次扩容延迟；代价是 directory 间接访问，以及更复杂的分裂和迭代状态维护。

## 8. 遍历

map 遍历顺序未定义，runtime 会随机化 slot 和 directory 的起始偏移。

语言语义要求：

- 已返回的 entry 不能重复返回。
- 迭代期间新增的 entry 可能返回，也可能不返回。
- 已修改 entry 必须返回最新值。
- 已删除 entry 不能返回。

迭代期间 table 增长或分裂时，迭代器继续扫描旧 table 来避免重复 key；对旧 table 选出的 key，再到新 table 查询最新 elem，并跳过已经删除的 key。directory 扩大时，迭代器还要调整目录索引。

`clearSeq` 用于识别迭代期间发生的 `clear`，随机起始偏移则避免形成稳定遍历顺序。

## 9. 编译器优化

### 9.1 make 与栈分配

逃逸分析使用 `EscNone` 表示 map 不逃逸。`src/cmd/compile/internal/walk/builtin.go:walkMakeMap` 会据此进行优化：

```text
不逃逸
→ Map 可以放在栈上
→ hint <= 8 时，首个 group 也可以预留在栈上
```

hint 为运行时值时，编译器可以生成条件分支：不超过 8 时使用预留 group，否则由 `runtime.makemap` 创建完整结构。后续增长产生的 directory、table 和 group 仍可能在堆上分配。

nil map 没有底层分配，因此不存在 `Map` 或 group 的栈堆判断。

### 9.2 字面量降低

编译器前端使用 `OMAPLIT` 表示 map 字面量。逃逸分析先处理 `OMAPLIT`，`src/cmd/compile/internal/walk/complit.go:maplit` 再生成 `OMAKEMAP` 并继承逃逸结果：

```text
OMAPLIT
→ escape analysis
→ maplit 生成 OMAKEMAP
→ OMAKEMAP 继承 OMAPLIT 的逃逸结果
→ walkMakeMap
→ 多次 map assignment
```

因此，不逃逸且条目数不超过 8 的字面量，同样可以使用栈上的 `Map` 和首个 group。map 本身不逃逸，不代表 key 或 elem 引用的对象也不逃逸。

### 9.3 静态与动态条目

对于运行时计算的条目：

```go
m := map[int]int{
	1:         10,
	loadKey(): loadValue(),
	3:         30,
}
```

`src/cmd/compile/internal/walk/order.go` 将动态条目拆出，保持动态表达式的求值和写入顺序。动态条目数量仍计入 `make` 的 hint，以减少初始化期间扩容。

### 9.4 大字面量

动态条目拆分后，静态条目不超过 25 个时直接生成逐项赋值。超过 25 个时，编译器生成只读 key/elem 数组和初始化循环，以减少机器码体积：

```go
for i := 0; i < len(staticKeys); i++ {
	m[staticKeys[i]] = staticElems[i]
}
```

每个元素仍经过正常哈希和写入路径，编译器不会预先构造依赖随机 `seed` 的 Swiss Table。

### 9.5 特化操作入口

编译器会根据 key 类型选择通用入口或 `fast32`、`fast64`、`faststr` 等特化入口，减少通用类型处理和间接调用成本。写入入口返回 elem slot 指针，随后由编译器生成实际 elem 赋值。

## 10. 并发与 GC

### 10.1 并发

普通 map 不提供并发同步：

- 多个 goroutine 只读同一个不再修改的 map 是安全的。
- 读写并发或多个写操作并发必须由调用方同步。
- `writing` 只能帮助 runtime 检测部分非法并发并终止程序，不能替代锁，也不能保证发现所有数据竞争。

需要并发读写时，应根据访问模式使用 `sync.RWMutex` 保护普通 map，或使用适合特定场景的 `sync.Map`。

### 10.2 GC 与对象生命周期

- map 会保持其中包含指针的 key 和 elem 可达，直到删除、`clear` 或 map 本身不可达。
- 删除时 runtime 会清零包含指针的存储，使 GC 能够回收不再引用的对象。
- key 和 elem 都不包含指针且没有因过大而间接存储时，group 数据不需要 GC 扫描，可以降低扫描成本。
- 大于 128 字节的 key 或 elem 使用间接存储，会增加独立分配和指针扫描。
- 包含指针的 key 写入 map 时，编译器需要保证指针目标不会因栈复制而移动，避免已保存哈希失效。

栈上的 `Map` 只减少顶层描述符和首个 group 的堆分配；map 增长后的内部结构以及 entry 引用的对象仍需分别判断分配位置和 GC 成本。

## 11. 与旧版实现对比

旧版 map 的核心实现集中在历史 `src/runtime/map.go`，使用 `hmap + bmap + overflow bucket`；当前实现改为 `Map + directory + table + group`。两者都在一个存储单元中放置 8 个 key/elem，并保存部分哈希以减少完整 key 比较，但解决冲突和增长的方式不同。

### 11.1 数据结构

```text
旧版
hmap
 └── buckets: 2^B × bmap
                    ├── tophash[8]
                    ├── 8 × key
                    ├── 8 × elem
                    └── overflow ──→ bmap ──→ ...

当前
Map
 └── directory: []*table
                    └── table
                         └── groups
                              ├── control word
                              └── 8 × slot
```

| 维度 | 旧版 bucket | 当前 Swiss Table |
|---|---|---|
| 顶层结构 | `hmap` | `maps.Map` |
| 元素数量 | `count` | `used` |
| 哈希种子 | `hash0` | `seed` |
| 基本单元 | `bmap`，8 个位置 | group，8 个 slot |
| 部分哈希 | `tophash[8]` | 8-byte control word 中的 H2 |
| 冲突处理 | overflow bucket 链 | group 开放寻址和三角数探测 |
| 增长状态 | `oldbuckets`、`nevacuate` | table 替换、分裂和 directory |

### 11.2 查找与冲突

旧版使用 hash 低 `B` 位选择主 bucket，再用高位 `tophash` 筛选 bucket 内的 8 个位置：

```text
hash 低位 → bucket
hash 高位 → 比较 tophash
→ 扫描当前 bmap
→ 未找到时沿 overflow 指针继续
```

当前实现使用哈希高位选择 directory 中的 table，使用 H1 探测 group，再用 H2 批量筛选 8 个 slot：

```text
hash 高位 → directory → table
H1 → 三角数探测 group
H2 → 批量匹配 control word
```

旧版冲突集中在同一个 bucket 链中，链变长会增加指针追踪和缓存未命中。当前实现不使用 overflow 链，冲突元素分散到探测序列中的其他 group，control word 可以在读取 key 前过滤大部分 slot；代价是高负载和 tombstone 会拉长探测序列。

### 11.3 删除状态

旧版 `tophash` 使用两种空状态：

- `emptyOne`：当前位置为空，但后面仍可能存在元素。
- `emptyRest`：当前位置及后续位置都为空，可以终止查找。

删除后 runtime 会尝试把 bucket 尾部连续的 `emptyOne` 收缩成 `emptyRest`。overflow 链本身保持不变。

当前 control byte 使用 `empty` 和 `deleted`：

- group 中已有 empty 时，删除位置可以变为 empty。
- group 没有 empty 时，删除位置必须变为 deleted，避免截断其他 key 的探测序列。

因此，两种实现都需要区分“可以终止查找的空位置”和“不能终止查找的删除位置”，只是旧版不变量作用于 bucket 及 overflow 链，当前实现作用于 group 探测序列。

### 11.4 扩容策略

旧版超过负载阈值时创建两倍 bucket 数组；overflow bucket 过多但元素负载不高时，还可以触发同容量增长来整理布局。迁移过程是增量的：

```text
buckets      → 新 bucket 数组
oldbuckets   → 保留旧数组
nevacuate    → 记录迁移进度
每次写入    → 搬迁命中的旧 bucket，并额外推进一个 bucket
```

增量迁移分散了整张 map 扩容的成本，但查找、写入和遍历都必须同时处理新旧 bucket。

当前实现让每个 table 独立增长：

```text
table 未达到上限 → 容量翻倍并重建该 table
table 达到上限   → 分裂成两个 table
目录位不足       → directory 翻倍
```

单次 table 重建不是逐 group 增量迁移，但 table 最大只有 1024 个 slot，因此迁移成本有上界。大型 map 只增长发生冲突的局部 table，不必重建全部数据。

### 11.5 遍历

两种实现都采用相同的核心策略处理迭代期间增长：继续扫描迭代开始时保存的旧存储，避免同一个 key 被返回两次；必要时到新存储重新查找，以获得最新 elem 或判断 key 是否已经删除。

- 旧版迭代器保存旧 bucket 数组，并识别 bucket 是否已经 evacuated。
- 当前迭代器保存旧 table，并处理 table 替换、分裂和 directory 扩大。

当前 directory/table 层次使迭代逻辑更复杂，但仍满足新增 entry 可以返回或跳过、修改必须返回最新值、删除不能返回的语义。

### 11.6 性能与 GC 权衡

| 维度 | 旧版 bucket | 当前 Swiss Table |
|---|---|---|
| 缓存局部性 | bucket 内较好，overflow 链需要指针追踪 | control word 和连续 group 更利于批量扫描 |
| 分支与比较 | 逐位置检查 `tophash` 和 key | 位运算批量筛选 H2，再比较候选 key |
| 冲突成本 | 分配和遍历 overflow bucket | 更长探测序列和 tombstone |
| 增长延迟 | 全局扩容，但增量 evacuation | 局部 table 整体重建，单次工作量有上界 |
| 额外内存 | overflow bucket、旧 bucket 数组 | directory、负载预留和 tombstone |
| 无指针数据 | 需要通过 `mapextra` 保持 noscan overflow 存活 | 无 overflow 链，满足条件的 group 可直接使用 noscan 分配 |

两种实现都只检测部分非法并发操作，不提供同步保证。Swiss Table 的主要收益来自减少 overflow 指针追踪、批量匹配 control word 和局部增长；代价是开放寻址不变量、tombstone、table 分裂及遍历逻辑更复杂。

## 12. 源码位置

- `src/internal/abi/map.go`：group 常量和 `MapType`。
- `src/internal/runtime/maps/map.go`：`Map`、小 map、directory 和顶层操作。
- `src/internal/runtime/maps/table.go`：查找、插入、删除、扩容、分裂和迭代。
- `src/internal/runtime/maps/group.go`：group、control word 和批量匹配。
- `src/internal/runtime/maps/runtime.go`：通用 runtime map 操作入口。
- `src/internal/runtime/maps/runtime_fast*.go`：整数、指针和字符串 key 的特化路径。
- `src/runtime/map.go`：编译器可见入口和创建逻辑。
- `src/cmd/compile/internal/walk/builtin.go`：`make(map)` 的编译期栈分配优化。
- `src/cmd/compile/internal/escape/expr.go`：map 创建与字面量的逃逸分析。
- `src/cmd/compile/internal/walk/order.go`：字面量动态条目的求值和写入顺序。
- `src/cmd/compile/internal/walk/complit.go`：字面量降低和大字面量循环初始化。
- 上游仓库历史 `src/runtime/map.go`：旧版 `hmap`、`bmap`、overflow bucket 和增量 evacuation。
