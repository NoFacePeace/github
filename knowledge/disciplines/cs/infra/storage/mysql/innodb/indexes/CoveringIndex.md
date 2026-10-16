---
color: "#5B6CDA"
---

# InnoDB 覆盖索引

## 1. 定义

当一个索引包含某条查询执行所需的全部列时，该索引对这条查询构成覆盖索引。它描述索引与查询之间的关系，不是一种独立的索引结构，也不需要专门的建索引语法。

同一个索引可以覆盖某条查询，却不能覆盖另一条查询。判断时需要考虑查询的输出、过滤、连接、排序和分组等操作使用的列。

## 2. 二级索引包含主键

InnoDB 二级索引记录包含索引列和聚簇索引键。有显式主键时，即使索引定义中没有写主键，记录中也包含主键列。

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    INDEX idx_name (name)
) ENGINE = InnoDB;
```

`idx_name` 包含 `name` 和 `id`，可以覆盖以下查询：

```sql
SELECT id, name FROM users WHERE name = 'Alice';
```

如果优化器使用 `idx_name`，通常无需为了读取所需列而访问聚簇索引。

## 3. 哪些查询不能被覆盖

对于上述 `idx_name`：

| 查询 | 是否覆盖 | 原因 |
| --- | --- | --- |
| `SELECT id FROM users WHERE name = 'Alice'` | 是 | 主键和过滤列都在索引中。 |
| `SELECT age FROM users WHERE name = 'Alice'` | 否 | 索引没有 age。 |
| `SELECT id FROM users WHERE name = 'Alice' AND age > 18` | 否 | 判断过滤条件需要 age。 |
| `SELECT id FROM users WHERE name = 'Alice' ORDER BY age` | 否 | 排序需要 age。 |
| `SELECT * FROM users WHERE name = 'Alice'` | 否 | 索引没有完整的表列。 |

这里判断的是索引是否包含查询所需数据，不表示优化器一定选择这个索引。

## 4. 联合索引实现覆盖

```sql
CREATE INDEX idx_name_age ON users (name, age);

SELECT id, name, age FROM users WHERE name = 'Alice';
```

`idx_name_age` 包含 `name`、`age` 和隐含主键 `id`，可以覆盖该查询。

覆盖能力与高效定位能力是两个问题：索引即使包含全部所需列，也可能需要扫描很多索引记录。联合索引列顺序仍应结合过滤与排序需求设计。

## 5. 与回表的关系及边界

覆盖索引通常可以减少为了获取缺失列而进行的[回表](TableLookup.md)，但并不保证运行时完全不访问聚簇索引。一致性读有时仍需访问聚簇索引确认记录可见性或读取历史版本。

前缀索引只包含字符串的一部分，不能仅凭前缀覆盖对完整字符串列值的读取。索引条件下推是在索引上提前检查适用条件，也不等于查询被索引覆盖。

## 6. 执行计划与维护成本

```sql
EXPLAIN SELECT id, name FROM users WHERE name = 'Alice';
```

可以关注 `key` 所选索引和 `Extra` 信息。`Using index` 通常表示通过覆盖索引获取查询所需列；`Using index condition` 表示索引条件下推，含义不同。

为了覆盖查询而增加索引列，会增加磁盘和缓存占用，以及相关列更新时的维护成本。应优先围绕常用查询评估收益，避免为了覆盖所有查询而建立过宽或重复的索引。
