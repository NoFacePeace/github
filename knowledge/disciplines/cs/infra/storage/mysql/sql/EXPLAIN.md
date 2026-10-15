---
color: "#5B6CDA"
---

# EXPLAIN

EXPLAIN 用于查看 MySQL 如何执行一条语句，包括表的访问方式、所选索引、连接与排序等操作。输出字段和操作含义见[执行计划](../server/ExecutionPlan.md)。

## 1. 基本用法

以下示例假设 users 表包含 id、name 和 age 列：

```sql
EXPLAIN
SELECT id, name FROM users WHERE age = 20;
```

普通 EXPLAIN 主要显示计划和估计值，不会像 EXPLAIN ANALYZE 一样完整执行目标查询来测量实际耗时。不同语句类型和版本支持范围有所不同，常用分析对象是 SELECT。

## 2. 输出格式

```sql
-- 显式指定传统表格格式。
EXPLAIN FORMAT=TRADITIONAL
SELECT id, name FROM users WHERE age = 20;

-- JSON 格式，提供更多计划与成本信息。
EXPLAIN FORMAT=JSON
SELECT id, name FROM users WHERE age = 20;

-- 树形格式，MySQL 8.0.16 起支持。
EXPLAIN FORMAT=TREE
SELECT id, name FROM users WHERE age = 20;
```

未显式指定 FORMAT 时，默认输出格式可能受版本和配置影响。

## 3. 读取结果

传统表格格式中，优先关注以下字段：

| 字段 | 检查内容 |
| --- | --- |
| table、type | 访问哪张表，是索引查找、范围扫描、索引合并还是全表扫描。 |
| possible_keys、key | 哪些索引可能用于定位，实际选择了哪个索引。 |
| key_len | 使用的索引键长度，辅助判断联合索引参与定位的部分。 |
| rows、filtered | 预计扫描行数及过滤后保留的比例。 |
| Extra | 是否覆盖索引、使用索引条件下推、额外排序或临时表。 |

不能只看到使用索引就判断查询高效，也不能只看到全表扫描就认定计划错误。需要结合数据规模、回表、排序和执行次数评估。

## 4. EXPLAIN ANALYZE

MySQL 8.0.18 起支持 EXPLAIN ANALYZE，它会实际执行查询，并报告执行节点的实际行数、耗时和循环次数。

```sql
EXPLAIN ANALYZE
SELECT id, name FROM users WHERE age = 20;
```

可重点检查估计行数与实际行数的差异，以及处理大量记录或被反复执行的节点。节点时间有包含关系，不能简单相加；多次循环时需结合 loops 理解统计值。

执行前应考虑查询自身的资源成本。EXPLAIN ANALYZE 的格式支持取决于版本，不应直接假设普通 EXPLAIN 的所有格式都适用。

## 5. 与 ANALYZE TABLE 的区别

- [ANALYZE](ANALYZE.md)：统计信息更新、直方图，以及与 EXPLAIN ANALYZE 的区别。

| 语句 | 作用 |
| --- | --- |
| EXPLAIN SELECT ... | 查看计划和估计值。 |
| EXPLAIN ANALYZE SELECT ... | 执行查询并显示实际执行统计。 |
| ANALYZE TABLE users | 更新表与索引统计信息，帮助优化器估算后续查询。 |

ANALYZE TABLE 不是执行计划查看命令，也不保证执行后查询一定变快。

[返回 SQL 目录](SQL.md)
