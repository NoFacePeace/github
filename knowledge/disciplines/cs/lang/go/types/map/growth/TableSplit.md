# Map 表拆分

表拆分将一个 table 的条目分配到两个新表，增加区分它们所需的哈希前缀位，并更新目录。本文按 Go 1.27.2 的 table.split 实现说明。

## 1. 触发条件

新键插入需要结构调整、清理未恢复预算，且原表容量翻倍会超过 maxTableCapacity 时，rehash 调用 split。该版本单表上限为 1024 个槽位。

拆分针对当前表，不要求把整个 map 的所有表一起重建。

## 2. 创建两个新表

split 将旧 localDepth 加一，创建左右两个新表。Go 1.27.2 中两表的容量均为 1024，而不是各分配旧容量的一半。

“拆分”指条目和哈希前缀范围分流，不表示把原组数组物理切成两半。两张新表拥有各自的组存储和增长预算。

## 3. 重新哈希并选择目标表

遍历旧表的有效条目，按类型规则取得实际键和值，再重新计算哈希：

```go
hash := typ.Hasher(key, m.seed)
if hash&mask == 0 {
    newTable = left
} else {
    newTable = right
}
newTable.uncheckedPutSlot(typ, hash, key, elem)
```

mask 由增加后的 localDepth 确定，检查从哈希最高位开始的第 localDepth 位。例如旧深度为 2，新深度为 3，就检查最高前缀的第 3 位。

```text
旧表对应前缀：10
    ├─ 新前缀 100 → 左表
    └─ 新前缀 101 → 右表
```

继续使用原 Map.seed，不主动更换种子。完整哈希不保存在槽的控制字节中，因此仍需调用哈希函数；对 NaN 等特殊键，也不能保证重算结果总与旧值相同。

## 4. 重新安排组和槽位

确定左右表后，迁移路径根据哈希在目标表内重新选择组和可用槽位，写入 H2 控制信息并更新对应表计数。

即使新表的组数与旧表相同，也需要重新插入，因为每张新表只承接部分条目，碰撞和探测布局可能不同。删除标记不随条目迁移，全 map 有效条目总数不因迁移改变。

## 5. 更新目录

安装新表时分两种情况：

| 原表 localDepth 与 globalDepth 的关系 | 目录处理 |
| --- | --- |
| 小于 globalDepth | 使用现有目录位数，更新相关目录项以区分左右表 |
| 等于 globalDepth | 先扩大目录，增加 globalDepth，再安装左右表 |

目录扩展需要同步维护 dirLen、dirPtr、globalShift 等状态。未拆分的其他表仍可被多个目录项共同引用，不需要因此重新计算其中全部条目的哈希。

## 6. 完成后的行为

旧表 index 被标为 -1，外层写入流程按新目录重新选择 table，继续尝试插入新键。条目分布不保证左右各一半，实际由哈希前缀决定。

迭代器可能保留旧表引用，旧存储不保证立即回收。表拆分和目录更新也不是业务并发同步机制。

## 7. 源码入口

- `src/internal/runtime/maps/table.go`：rehash、localDepthMask、split 和 uncheckedPutSlot。
- `src/internal/runtime/maps/map.go`：installTableSplit 与目录扩展。
