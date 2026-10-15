---
color: "#5B6CDA"
---

# MySQL 执行计划

## 1. 定义

执行计划是 MySQL 执行 SQL 的具体方案，由优化器选择，执行器按照计划完成操作。SQL 描述需要什么数据，执行计划决定如何获取和处理这些数据。

计划主要包括表的访问方式、所选索引、连接顺序与算法，以及过滤、排序、分组和聚合等操作。相同 SQL 在索引、数据分布、统计信息、配置或版本变化后，可能采用不同计划。

## 2. 查询示例

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    INDEX idx_age (age)
) ENGINE = InnoDB;

SELECT name FROM users WHERE age = 20;
```

可能的计划是通过 idx_age 找到年龄为 20 的索引记录，取得主键后回表读取 name，再返回结果。优化器也可能选择扫描全部表记录并过滤年龄，例如匹配行数很多、回表成本较高时。

## 3. 查看计划

```sql
-- 传统表格形式。
EXPLAIN SELECT name FROM users WHERE age = 20;

-- JSON 形式，提供更多计划和成本信息。
EXPLAIN FORMAT=JSON
SELECT name FROM users WHERE age = 20;

-- MySQL 8.0.16 起支持树形形式，便于查看操作层次。
EXPLAIN FORMAT=TREE
SELECT name FROM users WHERE age = 20;
```

普通 EXPLAIN 主要展示计划和估计值，不会像 EXPLAIN ANALYZE 一样完整执行目标查询来测量实际耗时。

## 4. 传统 EXPLAIN 字段

| 字段 | 含义 |
| --- | --- |
| id、select_type | 查询块的标识和类型，辅助识别子查询等结构。 |
| table | 当前访问的表或中间结果。 |
| type | 表的访问方式。 |
| possible_keys | 可能用于定位记录的索引；不等于所有可用于覆盖扫描的索引。 |
| key | 实际选择的索引；索引合并时可能列出多个。 |
| key_len | 使用的索引键长度，可辅助分析联合索引参与定位的部分。 |
| ref | 与索引比较的列、常量等信息。 |
| rows | 该次表访问预计扫描的行数，通常按每次访问估算。 |
| filtered | 预计经过该表条件过滤后保留的百分比。 |
| Extra | 覆盖索引、索引条件下推、排序等补充信息。 |

多表连接中的内层访问可能反复执行，因此不能简单把 rows 当作整条查询总共处理的行数。key_len 也不能单独反映索引列是否用于过滤或覆盖查询。

## 5. 常见访问方式与附加信息

### 5.1 访问方式

| type | 典型含义 |
| --- | --- |
| const | 通过主键或唯一非空索引的完整等值条件，最多匹配一行，并按常量表处理。 |
| eq_ref | 连接时，对前面表的每组记录，通过唯一键最多匹配一行。 |
| ref | 通过非唯一索引或唯一索引的部分前缀进行等值查找。 |
| range | 按一个或多个索引区间扫描。 |
| index_merge | 使用多个索引并合并匹配记录。 |
| index | 通常表示全索引扫描。 |
| ALL | 全表扫描；InnoDB 通常扫描聚簇索引。 |

不能仅按 type 判断查询是否高效。扫描规模、执行次数、回表与排序等共同决定成本。

### 5.2 Extra 信息

| 信息 | 典型含义 |
| --- | --- |
| Using index | 使用覆盖索引获取所需列，不表示索引定位一定高效。 |
| Using index condition | 使用索引条件下推，可能仍需回表。 |
| Using where | 需要进一步按条件过滤记录。 |
| Using filesort | 需要额外排序，不代表一定在磁盘上排序。 |
| Using temporary | 需要内部临时表，不代表一定使用磁盘存储。 |
| Using intersect / union / sort_union | 使用相应的索引合并算法。 |

覆盖查询在一致性读中仍可能为可见性判断而访问聚簇索引，不能把 Using index 理解为运行时绝不访问聚簇索引。

## 6. 实际执行信息

MySQL 8.0.18 起支持 EXPLAIN ANALYZE，它会实际执行查询，并显示操作的估计与实际行数、耗时和循环次数。

```sql
EXPLAIN ANALYZE
SELECT name FROM users WHERE age = 20;
```

诊断时应重点比较估计与实际行数的偏差，检查哪些节点处理了大量记录或被重复执行。节点时间包含层次关系，不能把各节点耗时直接相加；多次循环时还需要结合 loops 理解每次执行的统计值。

由于会执行查询，应在使用前考虑查询自身的资源开销。

## 7. 与 ANALYZE TABLE 的区别

| 命令 | 作用 |
| --- | --- |
| EXPLAIN | 查看计划及估计值。 |
| EXPLAIN ANALYZE | 实际执行查询，显示实际执行统计。 |
| ANALYZE TABLE | 更新表与索引统计信息，帮助后续优化，不执行指定的业务查询。 |

统计信息过旧可能导致估计偏差，但更新统计信息不保证计划一定变化或变快。还需结合数据分布、索引设计与实际访问成本判断。

[返回 Server 层](Server.md)
