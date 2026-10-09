---
color: "#5B6CDA"
---

# Quicklist

分块双向链表。将多个节点组织成双向链表，普通节点持有 listpack，大元素也可以存放在独立的普通数据节点中。

## 1. 在 Redis 中的用途

List 的底层存储结构。

## 2. 学习重点

- 外层节点链表与节点内部存储的关系。
- 两端操作、节点分裂与合并。
- 节点容量限制、内部节点压缩与两端节点保留不压缩的策略。

## 3. 源码入口

学习基线：Redis 7.2.16。

- [quicklist.h](https://github.com/redis/redis/blob/7.2.16/src/quicklist.h)
- [quicklist.c](https://github.com/redis/redis/blob/7.2.16/src/quicklist.c)

[返回数据结构目录](Data Structures.md)

