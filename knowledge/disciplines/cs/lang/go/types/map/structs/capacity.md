# table capacity

capacity 是 table 中的 uint16 字段，表示该表的总槽位数，包含当前占用、空闲和删除状态的槽位。本文按 Go 1.27.2 的实现说明。

## 1. 与组数的关系

每组包含 8 个逻辑键值槽，组数为 2 的幂，因此：

```text
capacity = 组数 × 8
         = (groups.lengthMask + 1) × 8
```

| 组数 | capacity |
| --- | --- |
| 1 | 8 |
| 2 | 16 |
| 8 | 64 |
| 128 | 1024 |

capacity 是槽位数，不是字节数。同样容量的不同 map 类型，可能因为键值大小、对齐或间接存储而占用不同内存。

## 2. 初始值不是固定的

未提供较大大小提示时，map 通常先采用小 map 单组模式，不存在独立的 table.capacity。该模式最多保存 8 个条目，插入第 9 个不同键时，通常转为容量为 16 的常规表。

使用 `make(map[K]V, hint)` 时，运行时根据提示、负载预算和容量约束估算初始空间，可能直接创建更大的表或多个表。hint 不直接等于某个 table 的 capacity，也不是条目数上限。

源码还支持容量为 8 的常规 table，但它与直接由 Map.dirPtr 指向 group 的小 map 模式不同，不能仅根据槽位数把二者混为一谈。

## 3. 总槽位不等于可插入额度

常规空表的增长预算通常为 capacity 的 7/8；最小单组 table 则保留一个空槽。有效元素数与删除标记会进一步消耗预算。

```text
capacity = 16
空表初始 growthLeft = 14
```

这不表示每组只能放 7 个条目，而是表级预算。空间仍有空槽时，也可能因为增长预算耗尽而需要清理、扩容或拆分。

## 4. 增长与上限

Go 1.27.2 的单表最大容量为 1024 个槽位。进入相应扩容路径时，通常将表容量翻倍；如果翻倍会超过单表上限，则拆分为两个表，配合目录更新区分条目。

```text
单表扩容：16 → 32 → 64 → … → 1024
达到单表上限后：按相应路径拆分
```

1024 是该版本单个 table 的上限，不是整个 map 的条目上限。一个 map 可以通过目录关联多个 table。

## 5. 与其他字段及 API 的区别

| 内容 | 含义 |
| --- | --- |
| table.capacity | 该表总槽位数 |
| table.used | 该表有效条目数 |
| table.growthLeft | 剩余增长预算 |
| Map.used，即 len(m) | 整个 map 的有效条目数 |
| Map.dirLen | table 指针目录的项数 |

Go 不支持 `cap(m)` 查询 map 容量，也没有将这个内部 capacity 暴露为稳定业务 API。

## 6. 删除与清空

删除条目一般不会自动减小 capacity，clear 也通常保留表存储以供复用。有效条目数降低，不代表槽位数组同步缩小。

需要减少保留空间时，可以按业务需求创建新 map 并复制剩余条目；旧存储是否可回收还取决于其他引用，GC 回收也不等于立即归还物理内存。

## 7. 源码入口

- `src/internal/runtime/maps/table.go`：capacity、newTable、reset、maxGrowthLeft、grow 和 split。
- `src/internal/runtime/maps/map.go`：大小提示、小 map 转换及目录组织。
