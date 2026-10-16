---
color: "#5B6CDA"
---

# IX 意向排他锁

IX（Intention Exclusive，意向排他锁）是 InnoDB 的表级锁，也是[意向锁](../../../../../runtime/locks/IntentionLock.md)的一种，表示事务正在或准备对表中的记录获取排他锁。它用于协调表级锁与记录锁，不表示事务独占整张表。

## 1. 获取场景

事务对记录获取 X 排他锁前，需要先在所属表上取得 IX 锁。常见操作包括 SELECT ... FOR UPDATE、INSERT、UPDATE 和 DELETE。

```sql
-- 假设 users 使用 InnoDB，id 是主键，且 id = 10 存在。
START TRANSACTION;

SELECT * FROM users WHERE id = 10 FOR UPDATE;

COMMIT;
```

此例中，事务先在 users 表上取得 IX，再对 id = 10 的记录获取 X 记录锁。事务已经持有适当的 IX 时，后续操作可以沿用，不需要每锁一行都重复获取表锁。

IX 由 InnoDB 自动管理，通常在事务提交或回滚时释放。普通一致性 SELECT 不会仅仅因为读取数据就获取 IX。

## 2. 兼容性

以下关系针对不同事务在同一张表上的锁：

| 另一事务持有的表级锁 | 与 IX 是否兼容 | 含义 |
| --- | --- | --- |
| IS 意向共享锁 | 兼容 | 两个事务可以分别声明记录共享锁和排他锁的意向。 |
| IX 意向排他锁 | 兼容 | 多个事务可以同时准备修改表中的记录。 |
| S 共享表锁 | 冲突 | 整表共享访问要求与表内记录的排他修改意向冲突。 |
| X 排他表锁 | 冲突 | 整表排他访问要求与其他事务的 IX 冲突。 |

表中的 S、X 指整表锁，不是记录锁。申请整表 S 或 X 时，可以通过检查已有意向锁发现冲突，无需遍历表内全部记录锁。

## 3. IX 兼容不代表记录锁兼容

假设事务 A、B 都通过主键等值查询执行 SELECT ... FOR UPDATE，且目标记录存在：

| 操作 | 表级锁 | 记录层面的结果 |
| --- | --- | --- |
| A 锁定 id = 10，B 锁定 id = 20 | 两个 IX 兼容。 | 对不同记录的 X 记录锁通常不冲突。 |
| A 锁定 id = 10，B 也锁定 id = 10 | 两个 IX 仍然兼容。 | 同一记录上的 X 锁冲突，B 需要等待 A 释放记录锁。 |

因此，判断一次操作是否等待，需要同时考虑表级锁、记录或范围锁，以及元数据锁等其他机制。不能只看到 IX 兼容就断定语句不会阻塞。

## 4. 与插入意向锁的区别

| 对比项 | IX 意向排他锁 | 插入意向锁 |
| --- | --- | --- |
| 作用层次 | 表级。 | 索引间隙级。 |
| 表达的意图 | 对表内记录获取排他锁。 | 在特定索引间隙位置插入记录。 |
| 主要协调对象 | 整表 S、X 锁。 | 阻止目标位置插入的间隙锁或临键锁中的间隙部分。 |

一次 INSERT 可以同时涉及表级 IX、插入意向锁以及新记录的排他保护。它们属于不同层次，不能因为名称都有“意向”就当作同一种锁。

## 5. 查看 IX 锁

MySQL 8.0 可通过 Performance Schema 查看，查询结果取决于权限和当前持锁状态：

```sql
SELECT ENGINE_TRANSACTION_ID, OBJECT_SCHEMA, OBJECT_NAME,
       LOCK_TYPE, LOCK_MODE, LOCK_STATUS
FROM performance_schema.data_locks
WHERE LOCK_TYPE = 'TABLE'
  AND LOCK_MODE = 'IX';
```

LOCK_STATUS 可用于区分已授予与正在等待的锁。看到表上的 IX 并不意味着整张表被独占，具体记录上的锁需要查看 RECORD 类型的锁信息。
