---
color: "#5B6CDA"
---

# MySQL 索引

索引用于组织数据并加速查询。MySQL 提供索引定义的 SQL 语法，存储引擎决定具体的存储结构、访问方式和支持范围。本文介绍索引分类及其与约束的关系。

## 1. 分类维度

以下维度可以重叠，每个维度文档链接对应的索引类型。

- [约束](constraints/Constraints.md)：主键索引、唯一索引和普通索引。
- [列数](column-count/ColumnCount.md)：单列索引和联合索引。
- [索引内容](contents/Contents.md)：前缀索引和函数索引。
- [特殊用途](special-purpose/SpecialPurpose.md)：全文索引和空间索引。

## 2. 索引结构

| 类型或结构 | 说明 |
| --- | --- |
| B+ 树索引 | InnoDB 常规主键、唯一和普通索引使用的结构，支持等值查找、范围查找及部分排序需求。 |
| HASH 索引 | 主要见于 MEMORY 引擎，适合等值查找，不支持按键的有序范围扫描。 |
| R-tree | InnoDB 空间索引使用的结构，利用空间对象的包围矩形进行检索。 |

InnoDB 的自适应哈希索引由引擎内部管理，不能像普通索引一样使用 SQL 手动创建。

## 3. 聚簇索引与二级索引

聚簇索引和二级索引描述的是存储组织方式，需要结合具体引擎理解。

- **InnoDB 聚簇索引**：叶子节点存储完整行，通常由主键构成，一张表只有一个。
- **InnoDB 二级索引**：叶子节点包含索引列和聚簇索引键，获取其他列时可能需要回表。

不能把 InnoDB 的组织方式推广到所有 MySQL 存储引擎。

## 4. 覆盖索引与分类重叠

覆盖索引表示某个索引包含了一条查询所需的全部列，是索引与查询之间的关系，不是一种独立的索引结构。

例如，InnoDB 表使用 `id` 作为主键时，`INDEX idx_name (name)` 的记录也包含 `id`，可以覆盖以下查询：

```sql
SELECT id, name FROM users WHERE name = 'Alice';
```

不同分类可以重叠。例如，InnoDB 中的 `UNIQUE INDEX uk_name_age (name, age)` 同时是唯一索引、联合索引和二级索引，采用 B+ 树结构；是否覆盖某条查询取决于该查询需要的列。

## 5. 查看与验证

```sql
-- 查看已有索引。
SHOW INDEX FROM users;

-- 查看表定义及索引、约束声明。
SHOW CREATE TABLE users;

-- 查看查询执行计划。
EXPLAIN SELECT id, name FROM users WHERE name = 'Alice';
```

建立索引不保证查询一定使用它。优化器会根据统计信息和访问成本选择计划，索引也会增加写入维护成本和存储开销。

[返回 MySQL 目录](../MySQL.md)
