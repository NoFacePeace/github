---
color: "#5B6CDA"
---

# ZSet

有序集合（Sorted Set），成员唯一，每个成员关联一个分数，按分数排序；分数相同时按成员字节的字典序排序。

## 1. 常见用途

排行榜、延迟任务。

## 2. 常用命令

`ZADD`、`ZREM`、`ZSCORE`、`ZINCRBY`、`ZRANGE`、`ZRANK`、`ZCARD`、`ZSCAN`。

## 3. 底层编码

listpack，或跳表与哈希表的组合；哈希表支持按成员查找分数，跳表支持排序和范围查询。

[返回数据类型目录](DataTypes.md)

