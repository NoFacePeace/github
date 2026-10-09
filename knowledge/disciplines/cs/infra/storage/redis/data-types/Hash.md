---
color: "#5B6CDA"
---

# Hash

字段与值的映射，一个键可以保存多个字段。

## 1. 常见用途

用户信息、对象属性。

## 2. 常用命令

`HSET`、`HGET`、`HMGET`、`HGETALL`、`HDEL`、`HEXISTS`、`HINCRBY`、`HSCAN`。

## 3. 底层编码

listpack 或哈希表，编码选择与字段数量、字段和值的大小有关。

实现说明以 Redis 7.2 为基线。

[返回数据类型目录](Data Types.md)

