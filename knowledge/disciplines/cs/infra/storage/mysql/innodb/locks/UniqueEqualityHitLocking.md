---
color: "#5B6CDA"
---

# 唯一索引等值查询加锁（命中）

在 InnoDB 的 REPEATABLE READ 下，锁定读取通过完整唯一键定位到一条已有记录时，通常只需要对匹配记录加锁，不保护其前方间隙。以下示例使用 SELECT ... FOR UPDATE，不适用于普通一致性 SELECT。

## 1. 主键等值查询

假设 users 使用 InnoDB，id 是主键，已有 id 为 10、20、30 的记录：

```sql
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
START TRANSACTION;

SELECT * FROM users WHERE id = 20 FOR UPDATE;

-- 后续业务操作。
COMMIT;
```

事务获取或沿用表级 IX，然后直接对聚簇索引 id = 20 申请 X 记录锁。这里不需要先取得临键锁，再释放间隙部分。

| 其他事务的操作 | 典型结果 |
| --- | --- |
| 更新或删除 id = 20 | 与 X 记录锁冲突，需要等待。 |
| 对 id = 20 执行 FOR SHARE 或 FOR UPDATE | 与已有记录锁冲突，需要等待。 |
| 插入 id = 15 或 25 | 不会仅因 id = 20 上的纯记录锁而等待。 |
| 普通一致性读取 id = 20 | 按 MVCC 判断版本可见性，通常不等待该记录锁。 |

## 2. 唯一二级索引

假设 email 为 NOT NULL 且具有唯一索引，执行以下查询并命中记录：

```sql
SELECT * FROM users WHERE email = 'alice@example.com' FOR UPDATE;
```

使用该唯一二级索引定位时，通常对匹配的二级索引记录加 X 记录锁，并根据其中的主键值锁定对应的聚簇索引记录。不会因此遍历所有主键。

## 3. 适用边界

联合唯一索引需要完整等值匹配所有组成列。只匹配其中一部分不能保证唯一，不能套用只锁一条记录的规则。

允许 NULL 的唯一索引可包含多条带 NULL 的记录，因此 IS NULL 查询不一定属于唯一记录查找。实际还需要确认优化器使用了预期索引，没有执行更大范围的扫描。

## 4. 生命周期与修改操作

相关事务锁通常在提交或回滚时释放。UPDATE、DELETE 使用相同唯一键定位时，搜索阶段也通常只需记录锁，但修改、删除二级索引项或执行约束检查可能涉及额外锁。
