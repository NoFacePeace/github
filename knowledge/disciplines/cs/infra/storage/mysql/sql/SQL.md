---
color: "#5B6CDA"
---

# SQL

SQL（Structured Query Language，结构化查询语言）用于定义数据结构、查询和修改数据，以及管理权限与事务。MySQL 在 SQL 标准的基础上提供自己的语法和功能，具体支持范围取决于版本。

## 1. 常见语句分类

| 分类 | 主要用途 | 常见语句 |
| --- | --- | --- |
| 数据定义（DDL） | 创建、修改和删除数据库对象。 | CREATE、ALTER、DROP、TRUNCATE |
| 数据操作（DML） | 插入、更新和删除记录。 | INSERT、UPDATE、DELETE |
| 数据查询（常单列为 DQL） | 查询、过滤、连接和聚合数据。 | SELECT |
| 数据控制（DCL） | 授予和撤销权限。 | GRANT、REVOKE |
| 事务控制（TCL） | 开始、提交和回滚事务。 | START TRANSACTION、COMMIT、ROLLBACK、SAVEPOINT |

上述分类是学习时的常用划分，部分资料将 SELECT 也归入 DML。

## 2. 查询示例

- [SELECT](SELECT.md)：查询语法、条件过滤、连接、分组聚合、排序与分页。

以下示例假设 users 表包含 id、name 和 age 列：

```sql
SELECT id, name
FROM users
WHERE age >= 18
ORDER BY id
LIMIT 10;
```

WHERE 筛选记录，ORDER BY 明确结果顺序，LIMIT 限制返回数量。SQL 的书写顺序不等于内部实际执行顺序，优化器会在保持语义的前提下选择执行计划。

## 3. 数据修改与事务

- [INSERT](INSERT.md)：插入单行与多行记录、查询结果写入、唯一键冲突处理及加锁行为。

以下示例假设 users 使用 InnoDB：

```sql
START TRANSACTION;

UPDATE users
SET age = age + 1
WHERE id = 10;

COMMIT;
```

事务提交前可以使用 ROLLBACK 撤销尚未提交的事务修改。事务能力取决于存储引擎；许多 DDL 语句会导致隐式提交，不能把它们当作普通 DML 一样通过用户事务回滚。

## 4. 与 MySQL 架构的关系

SQL 由 Server 层解析、优化并组织执行，存储引擎负责具体的数据读写。相同 SQL 的实际性能与表结构、索引、数据分布和执行计划有关。

## 5. 执行分析

- [EXPLAIN](EXPLAIN.md)：查看执行计划、选择输出格式，以及对比估计与实际执行信息。
- [ANALYZE](ANALYZE.md)：更新统计信息与直方图，区分 ANALYZE TABLE 和 EXPLAIN ANALYZE。

## 6. 管理语句

- [FLUSH](FLUSH.md)：刷新日志、表缓存、权限及状态信息，区分刷页、锁定与事务提交。
