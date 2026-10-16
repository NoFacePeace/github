---
color: "#5B6CDA"
---

# MySQL

MySQL 是关系型数据库管理系统，使用 SQL 进行数据查询与管理，支持事务、索引、复制等能力，常用于业务数据的持久化存储。

## 1. 架构

- [MySQL 架构](Architecture.md)：Server 层、InnoDB 存储引擎、日志机制与查询执行流程。

## 2. InnoDB

- [InnoDB](innodb/InnoDB.md)：默认存储引擎的数据与索引、内存缓存、事务与并发控制、日志与恢复机制。

## 3. 索引

- [MySQL 索引](indexes/Indexes.md)：索引分类、覆盖索引，以及索引与约束、存储引擎的关系。

## 4. SQL

- [SQL](sql/SQL.md)：语句分类、数据查询、数据修改与事务控制。

## 5. 日志

- [MySQL 日志](logs/Logs.md)：Server 层与 InnoDB 日志的分类、复制恢复及刷盘分工。

## 6. 事务

- [事务](../../../runtime/transactions/Transactions.md)：操作边界、提交与中止、并发控制及保证范围。

## 7. 性能

- [性能排查与优化](performance/Performance.md)：区分等待与执行开销，定位慢 SQL，并验证索引和查询优化效果。
