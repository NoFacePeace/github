---
color: "#5B6CDA"
---

# List

有序且允许重复的列表，支持两端插入和弹出。

## 1. 常见用途

队列、最近访问记录。

## 2. 常用命令

`LPUSH`、`RPUSH`、`LPOP`、`RPOP`、`BLPOP`、`BRPOP`、`LRANGE`、`LLEN`。

## 3. 底层编码

Redis 7.2 中主要使用 quicklist，节点通常以 listpack 存储元素。

实现说明以 Redis 7.2 为基线。

[返回数据类型目录](Data Types.md)

