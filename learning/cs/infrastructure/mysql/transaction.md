# MySQL 事务

先理解事务的整体语义，再认识并发问题，接着学习隔离级别提供的保证，最后深入 MVCC 与锁的实现机制。隔离级别属于事务的隔离性；理解其实现需要结合 MVCC 和锁。

## 1. 事务基础

- ACID、事务边界、自动提交、提交与回滚。
- 学习 `START TRANSACTION`、`COMMIT`、`ROLLBACK` 与 `SAVEPOINT`，明确事务开始、结束和部分回滚的位置。
- 以转账为例，理解多条 SQL 如何组成一个业务操作，以及约束和事务如何共同维护数据一致性。
- **实践**：执行转账事务，分别验证提交与回滚后的余额。

## 2. 并发问题

- **脏读**：读到其他事务尚未提交的修改。
- **不可重复读**：同一事务重复读取同一行，因其他事务提交修改而得到不同的值。
- **幻读**：同一事务重复按相同条件查询，因其他事务提交插入或删除而得到不同的行集合。
- **丢失更新**：两个事务基于同一个旧值计算并写回，后一次写入覆盖了前一次修改。例如，两个事务都读到计数值 100，各自在应用中加 10，再执行 `UPDATE counters SET value = 110 WHERE id = 1`，最终值为 110，而预期为 120。

前三项属于读异常，丢失更新属于写入冲突。分析丢失更新时，需要结合“读取 → 计算 → 写回”的完整流程、SQL 写法及并发控制方式，不能仅凭隔离级别名称判断是否安全。

## 3. 隔离级别

隔离级别规定并发事务之间的数据可见性。下面先按 SQL 标准对三类读异常的最低保证进行比较，“可能”表示该级别允许出现，并非一定发生。此表不涵盖丢失更新。

| 隔离级别 | 英文名称 | 脏读 | 不可重复读 | 幻读 |
| --- | --- | --- | --- | --- |
| 读未提交（RU） | READ UNCOMMITTED | 可能 | 可能 | 可能 |
| 读已提交（RC） | READ COMMITTED | 不允许 | 可能 | 可能 |
| 可重复读（RR） | REPEATABLE READ | 不允许 | 不允许 | 标准允许，具体行为见下文 |
| 串行化 | SERIALIZABLE | 不允许 | 不允许 | 不允许 |

**实现特点**：默认隔离级别为 RR。RC 的一致性读每次建立新快照；RR 的一致性读沿用同一事务内首次一致性读建立的快照，因此不会看到其他事务后来提交的新行。RR 的范围锁定读则通过间隙锁或 next-key lock 阻止向锁定范围插入。快照读与当前读的数据视图不同，混用时不能据此推断所有查询结果始终一致。详见[官方隔离级别文档](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html)。

## 4. 实践

**准备数据**：在本地学习数据库中执行一次，提交后再开始实验。两个会话均连接该数据库，并使用约定的默认存储引擎。

```sql
CREATE TABLE rr_demo (
    id INT PRIMARY KEY,
    value INT NOT NULL
);
INSERT INTO rr_demo VALUES (1, 100);
COMMIT;
```

**确认隔离级别**：两个会话均在事务外执行以下语句。默认值为 RR，但实验仍显式设置，避免已有配置影响结果。

```sql
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
SELECT @@session.transaction_isolation;
-- 预期：REPEATABLE-READ
```

### 4.1 RR：快照读与当前读

按表格从上到下执行，“—”表示该会话此步不操作。

| 步骤 | 会话 A | 会话 B | 预期结果 |
| --- | --- | --- | --- |
| 1 | `START TRANSACTION;` | — | A 开启事务，此时尚未建立一致性读快照。 |
| 2 | `SELECT value FROM rr_demo WHERE id = 1;` | — | 返回 100，首次一致性读建立快照。 |
| 3 | — | `START TRANSACTION;` | B 开启事务。 |
| 4 | — | `UPDATE rr_demo SET value = 120 WHERE id = 1;` | B 修改该行，尚未提交。 |
| 5 | `SELECT value FROM rr_demo WHERE id = 1;` | — | 仍返回 100，不读取 B 未提交的修改。 |
| 6 | — | `COMMIT;` | B 提交修改。 |
| 7 | `SELECT value FROM rr_demo WHERE id = 1;` | — | 仍返回 100，沿用原快照。 |
| 8 | `SELECT value FROM rr_demo WHERE id = 1 FOR UPDATE;` | — | 返回 120，当前读并加锁。 |
| 9 | `SELECT value FROM rr_demo WHERE id = 1;` | — | 仍返回 100，锁定读不会刷新原快照。 |
| 10 | `COMMIT;` | — | A 结束事务并释放锁。 |
| 11 | `START TRANSACTION;` | — | A 开启新事务。 |
| 12 | `SELECT value FROM rr_demo WHERE id = 1;` | — | 返回 120，新事务建立新快照。 |
| 13 | `COMMIT;` | — | 结束实验。 |

**观察重点**：RR 的可重复读针对同一事务内的一致性读；当前读不受旧快照限制。上述实验中 A 没有修改数据，如果 A 自己执行了更新，后续一致性读也能看到自身修改。普通 `START TRANSACTION` 不会立即建立快照；`START TRANSACTION WITH CONSISTENT SNAPSHOT` 可在 RR 下主动建立快照。参见[一致性非锁定读](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html)。

### 4.2 RR：快照建立时机

确认两个会话均已结束事务并使用 RR，将 `rr_demo` 中 `id = 1` 的值重置为 100 并提交，然后按以下顺序执行：

| 步骤 | 会话 A | 会话 B | 结果说明 |
| --- | --- | --- | --- |
| 1 | `START TRANSACTION;` | — | A 已开启事务，但尚未查询表数据，也未建立一致性读快照。 |
| 2 | — | `START TRANSACTION;` | B 开启事务。 |
| 3 | — | `UPDATE rr_demo SET value = 130 WHERE id = 1;` | B 将值更新为 130。 |
| 4 | — | `COMMIT;` | B 提交修改。 |
| 5 | `SELECT value FROM rr_demo WHERE id = 1;` | — | A 此时才建立快照，读到 B 已提交的 130。 |
| 6 | `COMMIT;` | — | A 结束事务。 |

**实测结果**：第 5 步返回 **130**。决定可见性的关键是快照建立时机，而非仅看事务开始时间：先查询建立快照，再发生其他事务的提交，后续普通查询沿用旧快照；先开启事务但不查询，等其他事务提交后才首次查询，则能读到该提交。

### 4.3 RC 与 RR 对比

确认两个会话均已结束事务，将数据重置为 100 并提交，把会话隔离级别改为 `READ COMMITTED` 后重做 4.1 的双会话实验。该实验第 7、9 步将返回 120，因为每次一致性读都会建立新快照。实验结束后可将会话隔离级别恢复为 `REPEATABLE READ`。
