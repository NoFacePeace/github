---
color: "#5B6CDA"
---

# ANALYZE

MySQL 中，ANALYZE 出现在两类用途不同的语句中：`ANALYZE TABLE` 更新统计信息，`EXPLAIN ANALYZE` 实际执行查询并报告执行统计。不能只根据 ANALYZE 这个词判断语句的作用。

## 1. ANALYZE TABLE

```sql
ANALYZE TABLE users;
```

此语句分析 users 表并更新表与索引统计信息，帮助优化器估算匹配行数及访问成本。InnoDB 通常通过采样索引页估算不同索引键值的数量，而不是精确统计每个值出现的次数。

它不修改业务行数据，也不是重建索引，但会更新统计信息，并可能影响之后的执行计划。

## 2. 什么时候考虑更新统计信息

大量新增或删除数据、数据分布明显改变，或执行计划中的估计行数与实际行数存在较大偏差时，可以检查统计信息是否需要更新。

InnoDB 在相关自动重算配置启用时，也会自动重新计算统计信息，因此不需要每次查询前手动执行 ANALYZE TABLE。

更新统计信息不保证查询一定变快。偏差还可能来自列之间的相关性、数据倾斜或不合适的索引设计。

## 3. 直方图

MySQL 8.0 支持通过以下语法显式收集列的直方图，帮助估算适用条件的选择性：

```sql
ANALYZE TABLE users
UPDATE HISTOGRAM ON age WITH 100 BUCKETS;
```

示例假设 age 列存在且其类型支持直方图。直方图保存值分布的统计摘要，不创建索引，也不能直接加速记录定位。

普通的 `ANALYZE TABLE users` 不等于执行 UPDATE HISTOGRAM；直方图的维护与更新行为还需结合具体版本及配置判断。

## 4. 执行成本

ANALYZE TABLE 会读取数据并消耗 I/O、CPU 等资源，执行中还可能涉及锁或锁等待，具体行为受存储引擎和版本影响。

大表应结合负载安排执行时机。统计信息变化也可能改变其他查询的执行计划，需要比较更新前后的实际效果。

## 5. EXPLAIN ANALYZE

MySQL 8.0.18 起支持 EXPLAIN ANALYZE，用于实际执行查询并展示估计与实际执行信息：

```sql
EXPLAIN ANALYZE
SELECT id, name FROM users WHERE age = 20;
```

示例假设 users 表包含 id、name 和 age 列。输出可以帮助检查实际返回行数、节点耗时和循环次数，判断估计偏差及主要开销。

EXPLAIN ANALYZE 不用于更新表的索引统计信息，并且会执行查询本身，不能把它当作纯粹查看计划的操作。

## 6. 命令对比

| 命令 | 主要作用 | 是否执行目标业务查询 |
| --- | --- | --- |
| EXPLAIN SELECT ... | 查看执行计划与估计值。 | 不完整执行查询以测量实际耗时。 |
| EXPLAIN ANALYZE SELECT ... | 显示实际执行统计。 | 是。 |
| ANALYZE TABLE users | 更新表与索引统计信息。 | 否，但会读取表或索引进行分析。 |
| ANALYZE TABLE ... UPDATE HISTOGRAM | 收集指定列的直方图。 | 否，但会读取数据进行统计。 |
