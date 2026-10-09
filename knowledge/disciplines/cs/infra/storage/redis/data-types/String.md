---
color: "#5B6CDA"
---

# String

字符串，可存储文本、数字或二进制数据。

## 1. 常见用途

缓存、计数器、分布式锁。

## 2. 常用命令

`SET`、`GET`、`MSET`、`MGET`、`INCR`、`DECR`、`APPEND`、`STRLEN`。

## 3. 底层编码

整数编码，或基于 SDS 的字符串编码。

实现说明以 Redis 7.2 为基线。

[返回数据类型目录](DataTypes.md)

