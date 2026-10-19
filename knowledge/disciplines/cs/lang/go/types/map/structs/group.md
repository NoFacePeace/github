# Map group

group 是 Go map 的组存储单元，包含控制信息和 8 个逻辑键值槽。groupsReference 定位整个组数组，其 group 方法返回指向其中一个组的 groupReference。本文按 Go 1.27.2 官方源码说明。

## 1. 单组引用

运行时使用以下结构定位一组数据：

```go
type groupReference struct {
    data unsafe.Pointer
}
```

groupReference 只保存地址，不直接包含 8 个键值。实际存储布局由编译器根据 map 的键值类型生成，运行时通过类型元数据访问。

## 2. 控制信息与槽位

```text
group
    ├─ 控制字：8 个控制字节，共 64 位
    └─ 8 个逻辑槽位：每槽关联一个键和一个值
```

第 i 个控制字节描述第 i 个槽位。控制信息紧凑，便于先批量筛选候选位置，再访问完整键值。

| 控制字节 | 含义 |
| --- | --- |
| `0x80` | 空槽 |
| `0xFE` | 删除槽 |
| `0x00` 至 `0x7F` | 占用槽，保存哈希低 7 位指纹 H2 |

控制信息不是完整哈希值。指纹匹配后仍必须比较完整键，才能确认查找命中。

## 3. 键值布局

Go 1.27.2 源码根据 mapsplitgroup 实验配置描述两种布局，以下是示意，不是可直接用于所有 map 的统一结构定义：

```text
交错布局：
    控制字 + [(key0, value0), (key1, value1), …]

分离布局：
    控制字 + [key0, key1, …] + [value0, value1, …]
```

类型元数据通过键值偏移和步长描述访问方式。较大键或值可能间接存储，实际组大小也会受对齐影响，不能简单按 8 倍键值大小推断全部内存占用。

### 3.1 交错布局的结构示意

未启用 mapsplitgroup 时，源码描述的布局可写作：

```go
// KeyType、ValueType 是示意占位，不是可直接编译的通用类型。
type group struct {
    ctrls ctrlGroup // 底层为 uint64。
    slots [8]struct {
        key  KeyType
        elem ValueType
    }
}
```

每个槽将键和值放在一起，访问某个条目时通常具有较好的局部性。字段对齐可能产生填充，因此实际大小需要依据类型布局计算。

### 3.2 分离布局的结构示意

启用 mapsplitgroup 时，键和值分别组织成数组：

```go
type group struct {
    ctrls ctrlGroup
    keys  [8]KeyType
    elems [8]ValueType
}
```

keys[i] 与 elems[i] 仍构成同一个逻辑槽位，由第 i 个控制字节描述。布局改变不意味着增加一层哈希表，也不改变每组 8 个逻辑槽位的关系。

### 3.3 实际 group 与 groupReference

上述 group 描述实际承载控制字和键值数据的存储。编译器根据具体 map 类型及构建配置生成相应组类型，运行时通过 GroupSize、键值偏移和步长访问它，不存在一个适用于所有 map 的固定键值结构。

groupReference 则只是指向该存储的引用：

```text
groupReference.data → 实际 group
                       ├─ ctrls
                       └─ 键值存储：交错槽或分离数组
```

因此，`type groupReference struct { data unsafe.Pointer }` 不是 group 数据布局本身，只是运行时操作单组时使用的地址包装。

## 4. 从组数组定位

groupsReference.group 根据组索引 i 和类型元数据中的 GroupSize 计算：

```text
groupReference.data = groupsReference.data + i × GroupSize
```

组索引必须由调用方保证合法。取得单组引用后，运行时再按槽索引读取控制字节、键和值；组索引和槽索引是不同层级。

## 5. 查找与探测

运行时先在组内批量匹配目标 H2，再依次检查候选槽的完整键。全部候选都不相等时，依据空槽等信息决定返回不存在，还是继续探测其他组。

删除槽不能无条件作为查找终点，否则可能遗漏位于后续探测组中的键。组只是表内探测的一个单位，不保证相同初始组的所有键都能放在这一组中。

## 6. 单组模式与常规表

常规模式下，group 属于 table 的组数组，由目录和表探测路径定位。小 map 模式下，Map.dirPtr 可以直接指向一个 group，跳过 table 和目录。

两者都使用组级存储，但负载预算、删除标记和增长路径存在区别。不能因为一组有 8 个槽位，就认为常规表允许每组都长期填满而不需要保留查找空间。

## 7. 源码入口

- `src/internal/runtime/maps/group.go`：groupReference、控制信息及组内访问。
- `src/internal/runtime/maps/table.go`：跨组探测及表级操作。
- `src/internal/runtime/maps/map.go`：小 map 的单组操作路径。
- `src/internal/abi`：组类型布局元数据。
