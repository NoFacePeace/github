---
color: "#5B6CDA"
---

# Intset

整数集合。在连续内存中按升序存储不重复的整数，以紧凑布局减少内存开销。

## 1. 学习重点

- 二分查找。
- 插入和删除时的元素移动。
- 16、32、64 位存储宽度的升级与元素重排。

## 2. 源码入口

学习基线：Redis 7.2.16。

- [intset.h](https://github.com/redis/redis/blob/7.2.16/src/intset.h)
- [intset.c](https://github.com/redis/redis/blob/7.2.16/src/intset.c)

[返回数据结构目录](Data Structures.md)

