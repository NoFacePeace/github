---
color: "#5B6CDA"
---

# Dict

字典（哈希表）。通过哈希函数定位桶，使用链式结构处理哈希冲突，支持平均 O(1) 的查找、插入和删除。

## 1. 学习重点

- 桶、条目与哈希冲突处理。
- 负载因子与扩缩容。
- 渐进式 rehash：通过分批迁移条目分摊成本；迁移期间同时维护新旧两张表。

## 2. 源码入口

学习基线：Redis 7.2.16。

- [dict.h](https://github.com/redis/redis/blob/7.2.16/src/dict.h)
- [dict.c](https://github.com/redis/redis/blob/7.2.16/src/dict.c)

[返回数据结构目录](DataStructures.md)

