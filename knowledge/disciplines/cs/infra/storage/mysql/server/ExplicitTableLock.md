---
color: "#5B6CDA"
---

# MySQL 显式表锁

显式表锁通过 LOCK TABLES 在 Server 层控制会话对指定表的访问，通过 UNLOCK TABLES 释放。READ 和 WRITE 分别具有共享、排他的访问语义，但不能直接等同于 InnoDB 内部的同一把锁。

## 1. READ 与 WRITE

| 模式 | 对应思想 | 持锁会话 | 其他会话 |
| --- | --- | --- | --- |
| READ | [S 锁](../../../../runtime/locks/Shared.md)的共享思想。 | 可以读取，不能修改。 | 可以读取，写入需要等待。 |
| WRITE | [X 锁](../../../../runtime/locks/Exclusive.md)的排他思想。 | 可以读取和修改。 | 对该表的普通读取和修改通常需要等待。 |

这里介绍普通 READ、WRITE 模式；READ LOCAL 等变体存在额外规则，不应直接套用全部结论。

## 2. 基本用法

```sql
LOCK TABLES users READ;
SELECT * FROM users;
UNLOCK TABLES;
```

锁作用于持锁会话，并非只保护紧接着的一条语句。需要一次访问多张表时，应在同一条 LOCK TABLES 中列出；表别名也需要与实际访问一致。

持锁期间不能任意访问未锁定的普通表。重新执行 LOCK TABLES 会释放原有显式表锁并获取新的锁，连接断开也会释放相关锁。

## 3. 与 InnoDB 内部表锁的关系

- [InnoDB 表锁](../innodb/locks/TableLocks.md)：表级锁模式、兼容关系及与 Server 层的分工。

对 InnoDB 表执行 LOCK TABLES 时，可能同时涉及 Server 表锁与 InnoDB 内部表锁。典型条件是 innodb_table_locks 开启且 autocommit 为 0。

- **Server 表锁**：协调 SQL 层的会话访问，不限于 InnoDB。
- **InnoDB 内部表锁**：与引擎的 IS、IX 等锁模式协调，将整表访问要求纳入事务锁体系。

例如某事务完成 UPDATE 语句但尚未提交，仍可能持有 IX 和记录 X 锁。内部表级 X 锁会与 IX 冲突，从而等待相关事务结束。

实际等待可能先发生在元数据锁等层次，不能把上述逻辑关系理解为每次都严格按同一顺序在两层阻塞。MDL 是保护对象定义的另一套机制。

## 4. 事务与释放

LOCK TABLES 会隐式提交当前事务，不能在已有事务中随意添加它来增强保护。对于需要显式表锁的 InnoDB 操作，可以采用以下常见顺序：

```sql
SET autocommit = 0;
LOCK TABLES users WRITE;

UPDATE users SET age = age + 1 WHERE id = 10;

COMMIT;
UNLOCK TABLES;
SET autocommit = 1;
```

示例假设 users 是 InnoDB 表且含有相应列。实际使用应在异常路径上处理回滚、解锁并恢复会话配置。

COMMIT 通常释放 InnoDB 事务锁，但不会自动释放 Server 的显式表锁，因此仍需 UNLOCK TABLES。START TRANSACTION 会释放已有显式表锁，不应在 LOCK TABLES 后用它代替上述事务安排。

UNLOCK TABLES 在释放通过 LOCK TABLES 获取的表锁时还可能导致隐式提交，不能把解锁当作事务回滚。一般业务优先评估行级锁或原子更新，避免不必要地扩大锁范围。

## 5. 与记录锁的区别

Server 的 WRITE 表锁通常阻止其他会话的普通 SELECT；InnoDB 的 X 记录锁则通常不阻止普通 MVCC 一致性读。二者虽然都体现排他思想，但保护范围、访问入口和实现规则不同。

持有表级 IX 也不等于持有 WRITE 表锁：多个事务可以同时持有 IX，并在记录层面分别处理冲突。
