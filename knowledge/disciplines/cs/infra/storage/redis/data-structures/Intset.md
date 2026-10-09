---
color: "#5B6CDA"
---

# Intset

整数集合。在连续内存中按升序存储不重复的整数，以紧凑布局减少内存开销。

## 1. 在 Redis 中的用途

Set 的一种底层编码。

## 2. 学习重点

- 二分查找。
- 插入和删除时的元素移动。
- 16、32、64 位存储宽度升级；宽度升级与 Set 转换为其他编码是不同过程。

## 3. 源码入口

学习基线：Redis 7.2.16。

- [intset.h](https://github.com/redis/redis/blob/7.2.16/src/intset.h)
- [intset.c](https://github.com/redis/redis/blob/7.2.16/src/intset.c)

[返回数据结构目录](Data Structures.md)

