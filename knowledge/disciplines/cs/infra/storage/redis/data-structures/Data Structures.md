---
color: "#5B6CDA"
---

# Data Structures

数据结构决定 Redis 如何组织和存储数据，同一种结构可以被多个数据类型复用。本文记录 Redis 的具体实现，学习基线为 [Redis 7.2.16](https://github.com/redis/redis/tree/7.2.16)。

## 1. 结构总览

| 数据结构 | 核心特点 | 主要关联 |
| --- | --- | --- |
| [SDS](SDS.md)（动态字符串） | 保存二进制安全的字符串，并显式记录长度；除紧凑的 sdshdr5 外，其余头部还记录分配容量。 | String 的 raw、embstr 编码，以及键名、字段名等字符串数据。 |
| [Dict](Dict.md)（字典（哈希表）） | 通过哈希函数定位桶，使用链式结构处理哈希冲突，支持平均 O(1) 的查找、插入和删除。 | 数据库键空间、Hash 和 Set 的哈希表编码，以及 ZSet 的成员到分数映射。 |
| [Intset](Intset.md)（整数集合） | 在连续内存中按升序存储不重复的整数，以紧凑布局减少内存开销。 | Set 的一种底层编码。 |
| [Listpack](Listpack.md)（紧凑列表） | 在一段连续内存中编码整数和字符串，减少单独分配元素及保存指针的开销。 | Hash、Set、ZSet 的紧凑编码；quicklist 节点和 Stream 消息块的内部存储。 |
| [Quicklist](Quicklist.md)（分块双向链表） | 将多个节点组织成双向链表，普通节点持有 listpack，大元素也可以存放在独立的普通数据节点中。 | List 的底层存储结构。 |
| [Skiplist](Skiplist.md)（跳表） | 通过随机层高建立多层索引，支持期望 O(log n) 的查找、插入和删除；最坏情况下可退化为 O(n)。 | ZSet 的 skiplist 编码，与字典共同维护同一批成员。 |
| [Rax](Rax.md)（基数树） | 按键的字节序列建立索引，通过压缩单一路径节省节点空间，并支持有序遍历。 | Stream 的消息块索引，以及消费者组中的部分管理结构。 |

## 2. 数据类型与数据结构

数据类型描述用户可以执行的操作，数据结构描述这些操作如何实现。例如，小型 ZSet 可以采用 listpack；较大的 ZSet 则采用跳表与字典的组合，数据类型仍然是 ZSet。

参见 [数据类型](../data-types/Data Types.md)。

[返回 Redis 目录](../Redis.md)

