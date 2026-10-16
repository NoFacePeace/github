---
color: "#5B6CDA"
---

# IS 意向共享锁

IS（Intention Shared，意向共享锁）是 InnoDB 的表级锁，也是[意向锁](../../../../../runtime/locks/IntentionLock.md)的一种，表示事务正在或准备对表内记录获取 S 共享锁。它用于协调表级锁与记录锁，不会锁住整张表中的所有记录。

## 1. 获取场景

事务对记录获取 S 共享锁前，需要先在所属表上取得 IS，或已经持有满足要求的更强表级锁模式。典型操作是 SELECT ... FOR SHARE，MySQL 8.0 支持该语法。

```sql
-- 假设 users 使用 InnoDB，id 是主键，且 id = 10 存在。
START TRANSACTION;

SELECT * FROM users WHERE id = 10 FOR SHARE;

COMMIT;
```

在事务此前未持有该表相关锁的情况下，此例先获取 users 表上的 IS，再对 id = 10 获取 S 记录锁。其他事务可以共享读取这条记录，但若要对它获取 X 记录锁，则需要等待。

IS 由 InnoDB 自动管理，通常在事务提交或回滚时释放。普通一致性 SELECT 使用 MVCC，不会仅仅因为读取记录就获取 IS；SERIALIZABLE 等条件下，普通 SELECT 的加锁行为可能不同。

## 2. 兼容性

以下关系针对不同事务在同一张表上的锁：

| 另一事务持有的表级锁 | 与 IS 是否兼容 | 含义 |
| --- | --- | --- |
| IS 意向共享锁 | 兼容 | 多个事务可以同时声明对记录获取共享锁的意向。 |
| IX 意向排他锁 | 兼容 | 记录共享锁与排他锁的意向可以共存，具体记录冲突另行判断。 |
| S 共享表锁 | 兼容 | 整表共享访问与记录共享访问的意向兼容。 |
| X 排他表锁 | 冲突 | 整表排他访问需要等待其他事务释放 IS。 |

表中的 S、X 指整表锁，不是记录锁。通过检查意向锁，可以判断整表锁请求是否冲突，无需逐条检查表内记录上的锁。

## 3. 意向锁兼容与记录锁冲突

假设事务 A 已通过主键等值查询，对存在的 id = 10 获取 S 记录锁并持有表级 IS：

| 事务 B 的操作 | 表级锁关系 | 记录层面的结果 |
| --- | --- | --- |
| 对 id = 10 执行 FOR SHARE | IS 与 IS 兼容。 | 同一记录的 S 锁兼容。 |
| 对 id = 10 执行 FOR UPDATE | IS 与 IX 兼容。 | 同一记录的 S 与 X 冲突，B 等待。 |
| 对 id = 20 执行 FOR UPDATE，且该记录存在 | IS 与 IX 兼容。 | 不同记录上的锁通常不冲突。 |

意向锁兼容不代表整条语句一定不等待，还需要考虑具体记录锁、范围锁和元数据锁等机制。

## 4. 与 IX、插入意向锁的区别

| 锁 | 作用层次 | 表达的意图 |
| --- | --- | --- |
| IS 意向共享锁 | 表级。 | 对表内记录获取共享锁。 |
| IX 意向排他锁 | 表级。 | 对表内记录获取排他锁。 |
| 插入意向锁 | 索引间隙级。 | 在特定位置插入记录。 |

IS 与 IX 都是表级意向锁，二者互相兼容。插入意向锁用于插入位置的协调，与表级 IS、IX 属于不同层次。

## 5. 查看 IS 锁

MySQL 8.0 可通过 Performance Schema 查看，查询结果取决于权限和当前持锁状态：

```sql
SELECT ENGINE_TRANSACTION_ID, OBJECT_SCHEMA, OBJECT_NAME,
       LOCK_TYPE, LOCK_MODE, LOCK_STATUS
FROM performance_schema.data_locks
WHERE LOCK_TYPE = 'TABLE'
  AND LOCK_MODE = 'IS';
```

事务已经持有 IX 等满足要求的更强模式时，共享锁定读取不一定额外显示独立的 IS。判断记录的实际保护情况，还需查看 RECORD 类型的锁信息。
