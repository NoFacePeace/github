---
color: "#5B6CDA"
---

# Dict

字典（哈希表）。通过哈希函数定位桶，使用[拉链法](../../../../../data-structures/hash-table/SeparateChaining.md)处理哈希冲突，支持平均 O(1) 的查找、插入和删除。

## 1. 学习重点

- [结构](struct/Struct.md)：字典字段、桶数组、键值条目与拉链法。
- 桶、条目与哈希冲突处理。
- 负载因子与扩缩容。
- [rehash](Rehash.md)：通过分批迁移条目分摊扩缩容成本；迁移期间同时维护新旧两张表。

## 2. 源码入口

- `dict.h`
- `dict.c`
