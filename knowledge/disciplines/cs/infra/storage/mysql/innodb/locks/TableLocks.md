---
color: "#5B6CDA"
---

# InnoDB 表锁

InnoDB 不只有记录和范围锁，也有表级锁。表级描述锁的作用层次，不代表持锁事务一定独占整张表；是否冲突取决于锁模式。

## 1. 常见类型

| 类型或模式 | 作用 |
| --- | --- |
| [IS、IX 意向锁](../../../../../runtime/locks/IntentionLock.md) | 在表上声明下层记录加锁的意图，协调表级锁与记录锁。 |
| [S 锁](../../../../../runtime/locks/Shared.md)、[X 锁](../../../../../runtime/locks/Exclusive.md) | 对整张表提出共享或排他的访问要求，与表上的其他模式进行兼容性检查。 |
| AUTO-INC | 特定自增值分配模式下使用的特殊表级锁，行为取决于配置和语句类型。 |

普通 DML 通常获取表级意向锁及相关记录锁，不会因为修改一行就自动升级为整表 X 锁。InnoDB 不通过锁升级将大量记录锁自动转换成表锁。

## 2. 意向锁与兼容性

获取记录 S 锁前，需要表上的 IS 或满足要求的更强模式；获取记录 X 锁前，需要表上的 IX。同一事务在同一表内的多次操作可以沿用适当的意向锁。

| 已持有／申请 | IS | IX | S | X |
| --- | --- | --- | --- | --- |
| IS | 兼容 | 兼容 | 兼容 | 冲突 |
| IX | 兼容 | 兼容 | 冲突 | 冲突 |
| S | 兼容 | 冲突 | 兼容 | 冲突 |
| X | 冲突 | 冲突 | 冲突 | 冲突 |

表中关系针对不同事务在同一张表上的锁。两个 IX 可以共存，事务是否修改同一记录，再由记录锁判断。

其他事务请求整表 S 或 X 时，可以检查意向锁判断冲突，无需逐条遍历所有记录锁。

## 3. S／X 表锁与显式表锁

LOCK TABLES 是 Server 层语句。对 InnoDB 表执行时，在 innodb_table_locks 开启且 autocommit 为 0 等典型条件下，还可以获取引擎内部的 S 或 X 表锁。

```text
LOCK TABLES users WRITE
        ↓
Server 层协调表访问
        ↓
满足相关条件时，获取 InnoDB 内部 X 表锁
        ↓
与其他事务已有的 IS／IX 等模式判断冲突
```

例如事务 A 已执行 UPDATE，但尚未提交，仍持有 IX 和记录 X。事务 B 请求内部表级 X 时会与 A 的 IX 冲突，从而等待。实际执行也可能先在 MDL 等层次发生等待，这不是固定的等待顺序图。

Server 表锁与 InnoDB 表锁是两套机制：前者协调 SQL 层的访问，后者将整表要求纳入引擎事务锁体系。Server 的 WRITE 表锁通常阻止其他会话的普通读取，不能套用“MVCC 普通读不等待 X 记录锁”来推断其行为。

## 4. 自增锁

- [自增锁](AutoIncrementLock.md)：AUTO-INC 锁、三种分配模式、语句生命周期和自增值空缺。

AUTO-INC 用于某些自增值分配场景。不同 innodb_autoinc_lock_mode 模式会在表级锁和轻量同步之间作不同取舍，并非所有 INSERT 都获取 AUTO-INC 表锁。

需要该表级锁时，通常持续到语句结束，而非事务结束。它与 IS／IX 的模式兼容表不能简单混用，自增值也不保证连续无间隙。

## 5. 生命周期与事务边界

IS、IX 和普通事务级 S／X 表锁通常随事务结束释放；AUTO-INC 表锁通常按语句管理。

Server 显式表锁不会因为 COMMIT 就自动全部释放，仍需按 LOCK TABLES 的规则解锁。LOCK TABLES 还会隐式提交当前事务，因此不能仅根据引擎锁的生命周期安排整个 SQL 流程。

Server 的元数据锁保护对象定义，与这里的 InnoDB 表级锁不同。同一条语句可以同时涉及这两层锁。

## 6. 查看锁

MySQL 8.0 可通过 Performance Schema 查看引擎数据锁，具体可见性受权限影响：

```sql
SELECT ENGINE, OBJECT_SCHEMA, OBJECT_NAME,
       LOCK_TYPE, LOCK_MODE, LOCK_STATUS
FROM performance_schema.data_locks
WHERE LOCK_TYPE = 'TABLE';
```

此查询不能替代 Server 层锁诊断，MDL 等信息应结合 metadata_locks 和会话状态查看。

[返回 InnoDB 锁目录](Locks.md)
