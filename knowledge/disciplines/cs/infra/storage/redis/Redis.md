---
color: "#5B6CDA"
---

# Redis

Redis 是以内存为主要数据存储介质的键值存储系统，支持多种类型、持久化、复制和集群，常用于缓存、计数器、排行榜和消息流等场景。

## 1. 类型

- [类型](types/Types.md)：String、List、Hash、Set、ZSet 和 Stream 的语义、命令及底层编码。

## 2. 结构

- [结构](structs/Structs.md)：SDS、字典、整数集合、listpack、quicklist、跳表和基数树的作用与实现。

## 3. 可靠性

- [可靠性](reliability/Reliability.md)：数据持久化与故障后的恢复。

## 4. 配置

- [配置](config/Config.md)：配置文件、运行时配置，以及 AOF 开关与同步策略。

## 5. 事务

- [事务](transactions/Transactions.md)：命令排队与连续执行、WATCH 乐观锁，以及取消执行与回滚的区别。

## 6. Lua

- [Lua](lua/Lua.md)：服务端脚本、条件判断与数据修改，以及错误处理和回滚限制。

## 7. 场景

- [场景](scenarios/Scenarios.md)：缓存、计数、限流、消息流和分布式锁等常见用途。
