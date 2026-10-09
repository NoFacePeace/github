---
color: "#5B6CDA"
---

# Rax

基数树。按键的字节序列建立索引，通过压缩单一路径节省节点空间，并支持有序遍历。

## 1. 在 Redis 中的用途

Stream 的消息块索引，以及消费者组中的部分管理结构。

## 2. 学习重点

- 压缩节点的内存布局与路径匹配。
- 插入时的节点分裂，以及删除后的结构调整。
- 迭代器与有序遍历；消息块内容由 listpack 保存。

## 3. 源码入口

学习基线：Redis 7.2.16。

- [rax.h](https://github.com/redis/redis/blob/7.2.16/src/rax.h)
- [rax.c](https://github.com/redis/redis/blob/7.2.16/src/rax.c)

[返回数据结构目录](Data Structures.md)

