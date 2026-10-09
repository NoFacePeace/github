---
color: "#5B6CDA"
---

# Skiplist

跳表。通过随机层高建立多层索引，支持期望 O(log n) 的查找、插入和删除；最坏情况下可退化为 O(n)。

## 1. 学习重点

- 按分数排序，分数相同时按成员字节序排序。
- 每层前向指针与 span 跨度：跨度用于排名查询和按排名定位。
- 底层 backward 指针与 tail 指针支持倒序遍历。
- 插入、删除和分数更新时的连接与跨度维护。

## 2. 源码入口

学习基线：Redis 7.2.16。

- [server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)
- [t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)

[返回数据结构目录](Data Structures.md)

