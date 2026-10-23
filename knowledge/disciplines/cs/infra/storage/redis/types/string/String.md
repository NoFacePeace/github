---
color: "#5B6CDA"
---

# String

字符串，可存储文本、数字或二进制数据。

## 1. 常见用途

缓存、计数器、分布式锁。

## 2. 常用命令

`SET`、`GET`、`MSET`、`MGET`、`INCR`、`DECR`、`APPEND`、`STRLEN`。

命令用法、返回值和示例见 [Commands](Commands.md)。

## 3. 底层编码

整数编码，或基于 SDS 的字符串编码。
