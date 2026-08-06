# Go Map 创建与访问流程

## 一、核心结构

Go 1.28 的 map 使用 Swiss Table 实现，结构层级如下：

```text
Map
└── directory
    └── table
        └── groups
            └── group
                ├── control word
                ├── key slots
                └── elem slots
```

- Map
  - 保存 map 的总体状态
  - 保存随机哈希种子
  - 通过 dirPtr 找到 directory 或 small map 的 group
- directory
  - 本质是 table 指针数组
  - 不单独定义 directory 结构体
- table
  - 一个完整的 Swiss Table
  - 保存容量、装载状态和 group 数组
- group
  - 固定包含 8 个 slot
  - 包含一个 8 字节 control word
- slot
  - 保存一个 key
  - 保存一个 value
  - 对应一个 control byte

## 二、Map 创建过程

### 1. 源码入口

用户代码：

```go
m := make(map[int]string, hint)
```

编译器处理入口：

```text
src/cmd/compile/internal/walk/builtin.go
```

runtime 入口：

```text
src/runtime/map.go
src/internal/runtime/maps/map.go
```

### 2. 编译期判断是否逃逸

编译器先根据逃逸分析判断 map 是否为 EscNone。

#### 不逃逸

概念上会准备栈上的 Map 描述符：

```go
var mapTemp maps.Map
m := &mapTemp
```

如果 hint 不超过一个 group 的容量，编译器还可能准备栈上的 group：

```go
var groupTemp maps.Group
groupTemp.ctrl = abi.MapCtrlEmpty
mapTemp.dirPtr = unsafe.Pointer(&groupTemp)
```

#### 逃逸

编译器不会把栈上的 Map 传给 runtime，而是传入 nil：

```go
m := runtime.makemap(mapType, hint, nil)
```

runtime 随后通过 new(Map) 或 newobject 分配描述符。

### 3. 小容量 map

当 hint 不超过 8 时，runtime 的 NewMap 只初始化 Map 和 seed：

```text
Map
├── used = 0
├── seed = random
├── dirPtr = nil
└── dirLen = 0
```

小 map 延迟分配 group 的原因是：

- make 后不写入时不需要分配 group
- 可以减少无效内存分配
- 第一次写入时再创建 group

### 4. 大容量 map

当 hint 大于 8 时，runtime 会：

1. 根据 hint 计算目标容量
2. 计算 directory 大小
3. 将 directory 大小向上取整为 2 的幂
4. 设置 globalDepth 和 globalShift
5. 创建 directory
6. 为 directory 中的每个位置创建 table
7. 为 table 创建 groups
8. 初始化每个 group 的 control word

最终结构：

```text
Map
├── seed
├── dirPtr -> directory
├── dirLen
├── globalDepth
└── directory
    └── table
        └── groups
```

### 5. 第一次写入小 map

逃逸的小 map 或延迟初始化的小 map，在第一次写入时执行：

```text
Map.dirPtr == nil
    └── growToSmall
        ├── 创建一个 group
        ├── control word 设置为 empty
        └── Map.dirPtr 指向 group
```

当这个 group 的 8 个 slot 都被使用后：

```text
small group
    └── growToTable
        ├── 创建 table
        ├── 迁移已有 key/value
        ├── 创建 directory
        └── Map.dirPtr 改为指向 directory
```

### 6. Map 字面量展开

源码：

```go
m := map[int]string{
	1: "a",
	2: "b",
}
```

map 字面量不会直接生成一个静态 map 对象，编译器会先创建 map，再生成多次写入。

小型字面量大致展开为：

```go
m := make(map[int]string, 2)

var keyTemp int
var valueTemp string

keyTemp = 1
valueTemp = "a"
m[keyTemp] = valueTemp

keyTemp = 2
valueTemp = "b"
m[keyTemp] = valueTemp
```

编译器使用临时 key/value 变量，使 map 写入参数满足可寻址要求。

后续的 map 赋值还会继续展开为 mapassign 函数：

```go
*runtime.mapassign_fast64(mapType, m, keyTemp) = valueTemp
```

实际使用的函数取决于 key 类型：

```text
mapassign
mapassign_fast32
mapassign_fast64
mapassign_faststr
```

大型字面量会使用静态 key/value 数组和循环：

```go
var keys = [...]int{1, 2, 3}
var values = [...]string{"a", "b", "c"}

m := make(map[int]string, len(keys))

for i := range keys {
	m[keys[i]] = values[i]
}
```

当前 Go 源码中，元素数量超过 25 时会采用这种数组加循环的形式。

空字面量：

```go
m := map[int]string{}
```

大致展开为：

```go
m := make(map[int]string, 0)
```

不会生成 map 写入操作。

map 字面量的完整编译期和运行期关系：

```text
map literal
    ├── make(map, hint)
    ├── key/value 临时变量
    ├── 多次 m[key] = value
    └── mapassign*
        └── runtime 写入 table/group/slot
```

数组字面量可以生成静态初始化数据，但 map 的最终 slot 位置依赖运行时 seed、
哈希计算和探测序列，因此 key/value 仍然需要经过 map 写入逻辑。

相关源码：

- /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/complit.go
- /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/expr.go

## 三、Map 访问过程

以读取为例：

```go
value, ok := m[key]
```

编译器会根据 key 类型选择 mapaccess 函数：

```text
mapaccess1
mapaccess2
mapaccess_fast32
mapaccess_fast64
mapaccess_faststr
```

### 1. 计算 hash

runtime 使用 key 类型对应的哈希函数：

```go
hash := typ.Hasher(key, m.seed)
```

每个 map 都有独立的 seed，避免不同 map 使用完全相同的哈希分布。

### 2. 拆分 H1 和 H2

```text
hash
├── H1：高 57 位
└── H2：低 7 位
```

```go
H1 = hash >> 7
H2 = hash & 0x7f
```

- H1 用于生成 group 探测序列
- H2 写入 control word，用于快速筛选候选 slot

### 3. 找到 table

如果是 small map：

```text
dirLen == 0
    └── dirPtr 直接指向 group
```

如果是普通 map：

```go
index := hash >> globalShift
table := directory[index]
```

directory 本质是：

```text
[]*table
```

当 directory 只有一个 table 时，直接使用下标 0。

多个 directory 位置可以指向同一个 table：

```text
directory
├── [0] -> table A
├── [1] -> table A
├── [2] -> table B
└── [3] -> table C
```

### 4. 使用 H1 找到初始 group

table 的 group 数量是 2 的幂，因此可以使用 mask：

```go
groupIndex := H1 & table.groups.lengthMask
```

源码会构造探测序列：

```go
seq := makeProbeSeq(H1, table.groups.lengthMask)
```

### 5. 三角探测寻找 group

探测序列不是简单的加一，而是：

```text
+1、+2、+3、+4、...
```

例如从 group 2、mask 为 7 开始：

```text
2 -> 3 -> 5 -> 0 -> 4 -> 1 -> 7 -> 6
```

实现逻辑：

```go
seq.index++
seq.offset = (seq.offset + seq.index) & seq.mask
```

三角探测的作用：

- 避免线性探测形成连续聚集
- group 数量为 2 的幂时可以遍历所有 group
- 使用加法和位运算，计算成本较低

### 6. 在 group 内匹配 H2

每个 group 有一个 8 字节 control word：

```text
[slot0][slot1][slot2][slot3][slot4][slot5][slot6][slot7]
```

runtime 使用 H2 一次筛选 8 个 control byte：

```go
matches := group.ctrls().matchH2(H2)
```

得到的是 slot 位图，例如：

```text
matches = 01010010
```

表示 slot 1、slot 4、slot 6 可能匹配。

### 7. 比较完整 key

H2 只是哈希摘要，不代表完整 key 相同：

```text
H2 相同
    └── 只是候选 slot
        └── 仍然要比较完整 key
```

伪代码：

```go
for matches != 0 {
	slot := matches.first()

	if group.key(slot) == key {
		return group.elem(slot), true
	}

	matches = matches.removeFirst()
}
```

如果当前候选 slot 的 key 不同，就继续检查 matches 中的下一个 slot。

### 8. 判断查找是否结束

当前 group 没有 H2 匹配时：

```text
存在 empty
    └── 查找结束，key 不存在

不存在 empty
    └── 继续探测下一个 group
```

deleted 不能作为结束条件，因为后续 group 可能仍有目标 key。

### 9. 完整读取流程

```text
key
└── Hasher(key, seed)
    ├── directoryIndex
    │   └── 找到 table
    │
    ├── H1
    │   └── 生成 group 探测序列
    │       ├── 找到 group
    │       ├── H2 匹配 control word
    │       ├── 完整 key 比较
    │       │   ├── 相同：返回 value
    │       │   └── 不同：继续下一个候选 slot
    │       ├── 遇到 empty：返回不存在
    │       └── 没有 empty：探测下一个 group
    │
    └── 返回 value 和 ok
```

## 四、Map 赋值过程

源码：

```go
m[key] = value
```

编译器会把赋值转换为 mapassign 调用和 value slot 写入：

```go
slot := runtime.mapassign_fast64(mapType, m, key)
*slot = value
```

runtime 的赋值流程：

```text
1. 检查 map 是否为 nil
2. 检查是否存在并发写入
3. 计算 key 的 hash
4. 设置 writing 标记
5. 找到 small group 或 directory/table
6. 搜索已有 key
7. 已有 key：返回原 value slot
8. 新 key：选择可用 slot
9. 写入 key、value 和 H2
10. 更新 used 和 growthLeft
11. 清除 writing 标记
```

### 1. 覆盖已有 key

如果 H2 匹配并且完整 key 相等：

```text
已有 key
    └── 不增加 used
        └── 返回已有 elem slot
            └── 覆盖 value
```

### 2. 插入新 key

如果 key 不存在：

```text
新 key
    ├── 查找 empty slot
    ├── 写入 key
    ├── 写入 value
    ├── 写入 H2
    ├── used++
    └── growthLeft--
```

真实 runtime 会返回 value slot 地址，由编译器生成的后续代码完成 value 写入。

### 3. small map 赋值

small map 没有 directory：

```text
Map.dirLen == 0
    └── dirPtr 直接指向 group
```

group 写满后：

```text
small group
    └── growToTable
        ├── 创建 16-slot table
        ├── 迁移原 group 中的有效元素
        └── 重试当前 key
```

### 4. table 赋值

普通 table 中：

```text
1. directory 找到 table
2. H1 生成 group 探测序列
3. H2 筛选候选 slot
4. 比较完整 key
5. 写入或覆盖
```

如果 `growthLeft == 0`，当前写入会先触发 tombstone 清理或 table 扩容。

## 五、Map 删除过程

源码：

```go
delete(m, key)
```

编译器会根据 key 类型选择 mapdelete 函数：

```text
mapdelete
mapdelete_fast32
mapdelete_fast64
mapdelete_faststr
```

runtime 的删除流程：

```text
1. map 为 nil 或没有元素
   └── 直接返回
2. 计算 key 的 hash
3. 找到 small group 或 directory/table
4. 按 H1 探测 group
5. 用 H2 找候选 slot
6. 比较完整 key
7. 找到 key 后清理 key/value
8. 根据 group 状态写入 empty 或 deleted
9. 更新 used、growthLeft 和 tombstonePossible
```

### 1. small map 删除

small map 没有探测序列，删除后可以直接恢复为 empty：

```text
slot
├── 清空 key
├── 清空 value
└── control byte = empty
```

### 2. table 删除

如果目标 group 中仍然存在 empty：

```text
deleted key
    └── control byte = empty
        └── growthLeft++
```

如果目标 group 已经没有 empty：

```text
deleted key
    └── control byte = deleted
        └── tombstonePossible = true
```

不能在满 group 中直接写 empty，否则会截断其他 key 的探测路径。

### 3. tombstone 的后续处理

后续插入时：

```text
优先复用 tombstone
```

如果 `growthLeft == 0`：

```text
1. 尝试 pruneTombstones
2. 清理成功
   └── 恢复 empty 和 growthLeft
3. 清理效果不足
   └── grow 或 split
```

## 六、Map 扩容过程

### 1. 触发条件

每个 table 维护：

```text
used
capacity
growthLeft
```

普通 table 的最大平均装载率约为 7/8：

```text
used + tombstones
    接近 capacity × 7/8
        └── growthLeft == 0
```

删除产生的 tombstone 也会消耗可用容量，因此即使有效元素不多，也可能触发整理。

### 2. 先清理 tombstone

当 `growthLeft == 0`：

```text
1. 尝试 pruneTombstones
2. 清理后恢复空间
   └── 继续插入
3. 仍然没有空间
   └── rehash
```

### 3. table 翻倍

当 table 容量没有超过最大容量：

```text
旧 table
    └── 新 table，容量翻倍
```

runtime 会：

1. 创建新的 table 和 groups
2. 遍历旧 table 的所有有效 slot
3. 重新计算 key 的 hash
4. 按新 group 数量重新探测
5. 迁移 key/value
6. 清除 tombstone
7. 更新 directory 中的 table 指针

### 4. table split

table 达到最大容量后拆分：

```text
旧 table
├── left table
└── right table
```

runtime 根据更高一位 hash 将元素分配到 left 或 right。

### 5. directory 扩展

如果 split 时：

```text
table.localDepth == map.globalDepth
```

directory 会扩大为原来的两倍：

```text
directory 长度 × 2
globalDepth++
globalShift--
```

然后把旧 table 的 directory 索引分别替换为 left 和 right。

### 6. 当前复刻的扩容限制

当前 `mini/go/map.go` 已实现：

- small group 转 table
- table 容量翻倍
- 旧元素重新哈希并迁移
- table 达到上限后 split
- directory 扩展和 table 指针替换

为了保持示例可读，当前复刻暂未实现：

- 真实 MapType 元数据
- 间接 key/elem
- 并发写入检测
- 迭代期间扩容语义

## 七、与当前复刻代码的对应关系

当前 mini/go 中的实现位于：

```text
map.go
```

对应关系：

- MapAccessInt
  - 模拟 mapaccess1 或 mapaccess2
- MapInsertInt
  - 模拟 mapassign
- MapDeleteInt
  - 模拟 mapdelete
- mapPruneTombstonesInt
  - 模拟 pruneTombstones
- mapHashInt
  - 模拟类型专用 Hasher
- mapDirectoryIndex
  - 模拟 runtime 的 directoryIndex
- mapMakeProbeSequence
  - 模拟 makeProbeSeq
- mapNextProbeSequence
  - 模拟 probeSeq.next
- mapMatchH2
  - 用循环模拟 control word 的批量 H2 匹配
- mapAccessGroupInt
  - 模拟 small map 的 group 查找
- mapAccessTableInt
  - 模拟 table 内的 group 探测和查找
- mapGrowSmallInt
  - 模拟 small group 转 table
- mapGrowTableInt
  - 模拟 table 翻倍或 split
- mapSplitTableInt
  - 模拟 table split 和 directory 更新

## 八、无法用普通 Go 完整复刻的部分

- 真实编译器生成的 mapaccess 快速函数选择
- internal/abi.MapType 的真实类型元数据
- 各种 key 类型对应的真实哈希函数
- AMD64 SIMD intrinsic 的 control word 匹配
- 真实 runtime 的 unsafe slot 布局
- GC 类型指针信息和写屏障
- 并发读写检测
- 迭代期间的复杂语义

当前复刻已经实现 int key/value 场景下的创建、访问、赋值、删除、tombstone
清理、table grow 和 table split。它使用普通 Go 循环表达核心控制流，重点保持
源码中的结构关系和运行时步骤。

## 九、源码位置

- Map 定义
  - /Users/haotao.chen/Desktop/repositories/go/src/internal/runtime/maps/map.go
- table 查找和探测
  - /Users/haotao.chen/Desktop/repositories/go/src/internal/runtime/maps/table.go
- control word 和 group
  - /Users/haotao.chen/Desktop/repositories/go/src/internal/runtime/maps/group.go
- group slot 数量
  - /Users/haotao.chen/Desktop/repositories/go/src/internal/abi/map.go
- 编译器 make(map) 展开
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/builtin.go
- 编译器 map 访问展开
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/expr.go
