---
color: "#5B6CDA"
---

# MySQL Server 层锁

Server 层通过锁协调对象定义、表访问、备份及会话间协作。Server 层的锁不都属于表锁，InnoDB 内部也存在表级锁；管理层次和锁定范围是不同维度。

## 1. 常见类型

| 类型 | 保护对象或用途 | 获取与释放 |
| --- | --- | --- |
| 元数据锁（MDL） | 保护表等对象的定义，协调 DML 与 DDL。 | 访问对象时自动获取，按锁的生命周期释放。 |
| 显式表锁 | 限制其他会话对指定表的读写。 | LOCK TABLES、UNLOCK TABLES。 |
| 全局读锁 | 暂时限制实例范围内普通持久化表的修改等操作，建立备份一致性点。 | FLUSH TABLES WITH READ LOCK、UNLOCK TABLES。 |
| 备份锁 | 阻止影响备份文件一致性的部分操作，同时允许普通 InnoDB DML。 | LOCK INSTANCE FOR BACKUP、UNLOCK INSTANCE。 |
| 命名锁 | 通过自定义名称协调应用操作。 | GET_LOCK、RELEASE_LOCK。 |

## 2. 元数据锁

MDL 防止对象在被使用时发生不兼容的结构变化。例如 SELECT、UPDATE 等访问表时会自动获取相应元数据锁，ALTER TABLE 则需要取得相应的 DDL 锁。

```sql
-- 会话 A。
START TRANSACTION;
SELECT * FROM users WHERE id = 10;
-- 暂不提交。

-- 会话 B：可能等待会话 A 释放元数据锁。
ALTER TABLE users ADD COLUMN remark VARCHAR(100);

-- 会话 A：结束事务后，相关事务级 MDL 被释放。
COMMIT;
```

普通表访问获取的事务级 MDL 通常持续到事务结束，而不是查询返回结果后立即释放。即使 SELECT 使用 MVCC、不获取记录锁，也可能持有 MDL 并阻塞 DDL。

等待中的 DDL 还可能使后续表访问排队，形成阻塞链。MDL 存在多种模式，在线 DDL 也会在特定阶段申请更强的锁，不能简单理解为所有读写与 DDL 都互相阻塞。

## 3. 显式表锁

- [显式表锁](ExplicitTableLock.md)：LOCK TABLES 的 READ／WRITE 语义、事务边界及与 InnoDB 表锁的关系。

```sql
LOCK TABLES users READ;
SELECT * FROM users;
UNLOCK TABLES;
```

READ 锁允许其他会话读取，阻止其写入；WRITE 锁通常阻止其他会话访问该表。持锁会话本身也受到锁定表列表和模式的限制。

LOCK TABLES 会隐式提交当前事务，与事务及 InnoDB 内部锁的交互需要特别区分，不能把它当作行锁的直接替代。具体行为还受 autocommit、innodb_table_locks 等配置影响。

## 4. 全局读锁

```sql
FLUSH TABLES WITH READ LOCK;

-- 完成需要锁保护的操作后释放。
UNLOCK TABLES;
```

全局读锁通常阻止普通持久化表的数据修改、相关 DDL 和事务提交等操作，普通读取仍可进行；临时表等对象存在例外。

获取过程可能等待正在执行的操作，持有期间会影响写入，不能将其视作只读备份命令本身。备份是否需要全局读锁、持有多久，取决于备份方式及涉及的存储引擎。

## 5. 备份锁

MySQL 8.0 提供实例级备份锁，需要 BACKUP_ADMIN 权限：

```sql
LOCK INSTANCE FOR BACKUP;

-- 完成相关备份操作后释放。
UNLOCK INSTANCE;
```

备份锁限制可能影响文件一致性的部分 DDL 和文件操作，允许普通 InnoDB DML 继续，通常比全局读锁限制更小。

它本身不会创建事务一致性快照，也不会完成文件复制或日志恢复所需的工作，不能仅靠取得备份锁就认定备份完整可恢复。

## 6. 命名锁

```sql
-- 最多等待 5 秒；返回 1 表示获取成功。
SELECT GET_LOCK('report_generation', 5);

-- 应用只在确认获取成功后执行约定的互斥操作。

SELECT RELEASE_LOCK('report_generation');
```

命名锁是协作锁：只有使用同一个名称并遵守约定的会话才会协调，不会自动阻止其他会话修改业务表。

锁归会话持有，不随 COMMIT 或 ROLLBACK 自动释放；应显式释放，会话断开时也会释放。锁的范围是当前 MySQL 实例，不能直接作为跨实例的分布式锁。

## 7. 与存储引擎锁的区别

InnoDB 的记录锁、间隙锁、临键锁、意向锁和 AUTO-INC 锁属于存储引擎机制。其中意向锁与 AUTO-INC 锁是表级锁，因此不能简单按“Server 表锁、InnoDB 行锁”划分。

一次 SQL 可以同时涉及多个层次的锁。例如 UPDATE 通常先涉及表的 MDL，再在 InnoDB 中获取相应意向锁和记录相关的锁。

共享、排他描述锁模式，表、记录、间隙等描述保护对象或范围，二者也不能混为同一种分类。

## 8. 查看锁信息

```sql
-- 查看元数据锁；可见性与采集情况受权限和 Performance Schema 配置影响。
SELECT OBJECT_TYPE, OBJECT_SCHEMA, OBJECT_NAME,
       LOCK_TYPE, LOCK_DURATION, LOCK_STATUS, OWNER_THREAD_ID
FROM performance_schema.metadata_locks;

SHOW PROCESSLIST;
```

可结合锁状态和线程信息定位持有者、等待者及长事务。MySQL 8.0 的 performance_schema.data_locks、data_lock_waits 主要用于查看 InnoDB 等引擎暴露的数据锁，不等同于 MDL 信息。

Server 内部还使用 mutex、读写锁等同步原语保护内存结构，通常不与本文面向 SQL 操作的锁混在一起分类。
