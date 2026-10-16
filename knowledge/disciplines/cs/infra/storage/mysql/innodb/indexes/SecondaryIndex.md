---
color: "#5B6CDA"
---

# InnoDB 二级索引

## 1. 数据组织

除[聚簇索引](ClusteredIndex.md)之外的常规 [B+ 树](../../../../../data-structures/BPlusTree.md)索引称为二级索引。一张 InnoDB 表可以有多个二级索引，其叶子记录包含索引列和聚簇索引键，而不是完整行数据或固定的磁盘行地址。

有显式主键时，聚簇索引键就是主键。若二级索引已包含某个主键列，InnoDB 不必再重复附加该列。

## 2. 普通与唯一二级索引

在已定义主键的表中，普通索引和其他唯一索引通常都属于二级索引：

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200) NOT NULL,
    age INT NOT NULL,
    INDEX idx_name (name),
    UNIQUE INDEX uk_email (email)
) ENGINE = InnoDB;
```

`idx_name` 允许重复的姓名，`uk_email` 检查邮箱是否重复；两者的叶子记录都包含主键 `id`。唯一性检查针对声明的唯一键，不能因为记录中附带主键，就认为不同主键允许相同的唯一键值。

## 3. 回表

- [回表](TableLookup.md)：执行流程、主键隐含存储、覆盖查询的边界与执行计划判断。

当查询需要二级索引中未包含的列时，通常先查二级索引取得主键，再到聚簇索引查找完整记录，这个过程称为回表。

```sql
SELECT age FROM users WHERE name = 'Alice';
```

如果优化器选择 `idx_name`，执行过程通常为：

1. 在二级索引中定位姓名为 `Alice` 的记录。
2. 从索引记录取得主键 `id`。
3. 使用 `id` 访问聚簇索引，读取 `age`。

多个匹配记录可能产生多次聚簇索引访问，实际磁盘读取量取决于缓存和执行策略。

## 4. 覆盖索引

- [覆盖索引](CoveringIndex.md)：判定方式、隐含主键、联合索引覆盖能力与使用边界。

```sql
SELECT id, name FROM users WHERE name = 'Alice';
```

`idx_name` 已包含所需的 `name` 和 `id`，因此可以覆盖该查询，通常减少回表。覆盖索引是针对查询而言的，不是独立的索引结构。

一致性读在需要确认记录可见性等情况下，仍可能访问聚簇索引，因此覆盖查询不等于任何情况下都不会访问聚簇索引。

## 5. 维护成本与执行计划

二级索引占用额外的磁盘和缓存空间。插入、删除、修改索引列或主键时，需要维护相关索引。

索引存在不保证查询一定使用它；当匹配记录很多且需要大量回表时，优化器可能选择扫描聚簇索引。应结合统计信息和执行计划判断索引效果。

```sql
EXPLAIN SELECT age FROM users WHERE name = 'Alice';
```

[返回 InnoDB 索引](Indexes.md)
