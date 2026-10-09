---
color: "#5B6CDA"
---

# Set

成员唯一的无序集合，支持交集、并集和差集。

## 1. 常见用途

去重、标签集合、共同关注。

## 2. 常用命令

`SADD`、`SREM`、`SISMEMBER`、`SCARD`、`SINTER`、`SUNION`、`SDIFF`、`SSCAN`。

## 3. 底层编码

Redis 7.2 中可使用 intset、listpack 或哈希表，编码选择与成员内容、数量和大小有关。

实现说明以 Redis 7.2 为基线。

[返回数据类型目录](Data Types.md)

