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

## 2. 结构笔记

- [zskiplist](zskiplist.md)：跳表整体结构、字段与节点连接关系。
- [zskiplistNode](zskiplistNode.md)：节点字段、层高、前向指针与跨度。
- [zskiplistLevel](zskiplistLevel.md)：节点在每层的连接结构，包含 forward 和 span。

## 3. 源码入口

- `server.h`
- `t_zset.c`

[返回数据结构目录](../DataStructures.md)

