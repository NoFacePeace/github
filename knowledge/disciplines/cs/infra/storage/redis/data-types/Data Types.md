---
color: "#5B6CDA"
---

# Data Types

Redis 数据类型定义一个键可以存储的内容和支持的操作；底层编码决定数据如何存储，同一种数据类型可以采用不同编码。

## 1. 核心数据类型

- [String](String.md)：字符串，可存储文本、数字或二进制数据。
- [List](List.md)：有序且允许重复的列表，支持两端插入和弹出。
- [Hash](Hash.md)：字段与值的映射，一个键可以保存多个字段。
- [Set](Set.md)：成员唯一的无序集合，支持交集、并集和差集。
- [ZSet](ZSet.md)：有序集合（Sorted Set），成员唯一，每个成员关联一个分数，按分数排序；分数相同时按成员字节的字典序排序。
- [Stream](Stream.md)：按消息 ID 组织的追加式消息流，支持消费者组和消息确认。

## 2. 扩展功能

Bitmap 和 HyperLogLog 基于 String，Geo 基于 ZSet。这些功能提供专门的命令，但不是独立的核心对象类型。

## 3. 类型与编码

- 使用 `TYPE key` 查看键的数据类型。
- 使用 `OBJECT ENCODING key` 查看键当前的底层编码。
- 编码会受 Redis 版本、配置和数据规模影响；本文的实现说明以 Redis 7.2 为基线。

[返回 Redis 目录](../Redis.md)

