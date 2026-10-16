---
color: "#5B6CDA"
---

# InnoDB 回表

## 1. 定义

回表是指通过二级索引取得聚簇索引键后，再访问聚簇索引获取记录。最常见的原因是查询需要二级索引中未包含的列。

有显式主键时，聚簇索引键就是主键。回表不一定读取磁盘：所需页已经在 Buffer Pool 中时，可以从内存读取。

## 2. 执行流程

以下示例使用主键 `id` 和二级索引 `idx_name`：

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    INDEX idx_name (name)
) ENGINE = InnoDB;

SELECT age FROM users WHERE name = 'Alice';
```

如果优化器选择 `idx_name`，通常按以下流程读取：

1. 在二级索引中定位 `name = 'Alice'` 的记录。
2. 从二级索引记录取得主键 `id`。
3. 根据 `id` 访问聚簇索引，读取 `age`。

多个匹配记录可能需要多次聚簇索引访问。若优化器直接扫描聚簇索引，则没有“先查二级索引再回表”的过程。

## 3. 查询索引列和主键是否回表

InnoDB 二级索引记录会包含聚簇索引键，即使建索引时没有显式写入主键列。

因此，示例中的 `INDEX idx_name (name)` 实际也保存了主键 `id`，可以覆盖以下查询：

```sql
SELECT id, name FROM users WHERE name = 'Alice';
```

如果使用 `idx_name`，通常无需为了获取 `id` 和 `name` 回表。联合主键的各个列也会包含在二级索引记录中；已经显式包含的主键列无需重复附加。

## 4. 判断是否需要回表

应检查查询执行所需的全部列，不只看 SELECT 列表，还包括过滤、连接、排序及分组等操作使用的列。

| 查询 | 使用 idx_name 时的典型情况 |
| --- | --- |
| `SELECT id, name FROM users WHERE name = 'Alice'` | 索引覆盖所需列，通常无需回表。 |
| `SELECT age FROM users WHERE name = 'Alice'` | 需要回表读取 age。 |
| `SELECT id FROM users WHERE name = 'Alice' AND age > 18` | 需要读取 age 判断条件，通常需要回表。 |
| `SELECT * FROM users WHERE name = 'Alice'` | 索引未包含所有列，通常需要回表。 |

一致性读有时需要访问聚簇索引确认记录可见性或读取历史版本，因此覆盖查询不保证任何情况下都不会访问聚簇索引。

## 5. 成本与优化

回表成本受匹配记录数量、聚簇索引页的分布和缓存命中率影响。大量匹配加上大量回表时，优化器可能认为直接扫描聚簇索引更划算。

可以通过更有选择性的条件减少候选记录，或围绕常用查询设计覆盖索引，减少为获取列值而进行的回表。增加索引列也会带来空间和写入维护成本。

索引条件下推可以在索引记录上提前检查适用条件，减少需要回表的候选记录，但不会把索引中未包含的列变为可直接读取的列。

## 6. 查看执行计划

```sql
EXPLAIN SELECT id, name FROM users WHERE name = 'Alice';
EXPLAIN SELECT age FROM users WHERE name = 'Alice';
```

可以关注 `key` 所选索引和 `Extra` 信息。`Using index` 通常表示使用覆盖索引；`Using index condition` 表示索引条件下推，两者含义不同。不能仅凭这些标记断言运行时绝不会为可见性检查访问聚簇索引。

[返回二级索引](SecondaryIndex.md)
