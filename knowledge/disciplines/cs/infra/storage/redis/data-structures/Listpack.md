---
color: "#5B6CDA"
---

# Listpack

紧凑列表。在一段连续内存中编码整数和字符串，减少单独分配元素及保存指针的开销。

## 1. 学习重点

- 头部、元素编码和结束标记。
- 元素自身的长度信息与反向遍历。
- 插入、删除导致的内存移动，以及紧凑存储与操作成本的取舍。

## 2. 源码入口

学习基线：Redis 7.2.16。

- [listpack.h](https://github.com/redis/redis/blob/7.2.16/src/listpack.h)
- [listpack.c](https://github.com/redis/redis/blob/7.2.16/src/listpack.c)

[返回数据结构目录](DataStructures.md)

