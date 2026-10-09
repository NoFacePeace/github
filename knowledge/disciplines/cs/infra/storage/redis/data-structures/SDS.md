---
color: "#5B6CDA"
---

# SDS

动态字符串。保存二进制安全的字符串，并显式记录长度；除紧凑的 sdshdr5 外，其余头部还记录分配容量。

## 1. 学习重点

- 不同头部类型与内存布局。
- 长度、容量与扩容策略。
- 二进制安全与末尾空字符的作用。

## 2. 源码入口

学习基线：Redis 7.2.16。

- [sds.h](https://github.com/redis/redis/blob/7.2.16/src/sds.h)
- [sds.c](https://github.com/redis/redis/blob/7.2.16/src/sds.c)

[返回数据结构目录](Data Structures.md)

