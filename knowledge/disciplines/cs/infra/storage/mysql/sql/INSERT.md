---
color: "#5B6CDA"
---

# INSERT

INSERT 用于向表中新增记录，可以插入显式指定的值，也可以将查询结果写入目标表。在 InnoDB 中，插入受事务、唯一约束、外键及并发锁控制。

## 1. 基本用法

以下示例假设 users 使用 InnoDB，包含自增主键 id、具有唯一约束的 name，以及 age 列，其他列均允许省略：

```sql
INSERT INTO users (name, age)
VALUES ('Alice', 20);
```

列清单与 VALUES 中的值按位置对应。显式指定列名可以避免依赖表中列的定义顺序，也能明确每个值写入哪个字段。

省略的列按列定义使用默认值、自增值或 NULL；不允许为空且没有可用默认值的列，在严格模式下通常会导致错误。不要依赖非严格模式下的隐式转换或补值行为。

## 2. 多行插入

```sql
INSERT INTO users (name, age)
VALUES ('Bob', 21),
       ('Carol', 22),
       ('David', 23);
```

一次写入多行可以减少网络往返与语句处理开销。批次过大会增加事务持锁时间、日志量和内存消耗，还可能受 max_allowed_packet 等限制，应结合记录大小和业务负载选择批次。

对 InnoDB，普通 INSERT 因重复键等错误失败时，通常回滚该语句的修改；这不等于自动回滚事务中此前成功的所有语句。死锁等错误可能导致整个事务回滚，应用需按实际错误处理事务状态。

## 3. 插入查询结果

假设 archived_users 表包含兼容的 name 和 age 列：

```sql
INSERT INTO archived_users (name, age)
SELECT name, age
FROM users
WHERE age >= 60;
```

查询输出的列数和顺序需要与目标列清单对应，值也必须满足目标表的数据类型及约束。目标表的唯一键冲突仍会影响插入结果。

INSERT ... SELECT 同时涉及源表读取和目标表写入，源表的读取与加锁行为受隔离级别等因素影响，不能直接套用普通独立 SELECT 的快照读规则。

## 4. 唯一键冲突处理

普通 INSERT 遇到主键或唯一索引冲突时会报错。应根据业务需要选择冲突处理方式：

| 语法 | 行为 |
| --- | --- |
| INSERT | 遇到重复键通常报错，使调用方处理冲突。 |
| INSERT IGNORE | 重复键对应的行通常被跳过，语句可以继续处理其他行；部分错误会转为警告或发生值调整。 |
| INSERT ... ON DUPLICATE KEY UPDATE | 插入发生主键或唯一键冲突时，转为更新已有记录。 |

```sql
INSERT INTO users (name, age)
VALUES ('Alice', 25)
ON DUPLICATE KEY UPDATE age = 25;
```

此例中，如果 name 为 Alice 的记录已经存在，就更新其 age；没有冲突则插入新记录。更新过程仍需满足其他约束，多条唯一索引可能使冲突处理更复杂。

INSERT IGNORE 不仅忽略重复键错误，还可能掩盖数据问题，执行后应检查警告。REPLACE 则可能先删除冲突记录再插入，并非普通 UPDATE 的替代写法。

## 5. 事务与自增值

```sql
START TRANSACTION;

INSERT INTO users (name, age)
VALUES ('Eve', 24);

SELECT LAST_INSERT_ID();

COMMIT;
```

LAST_INSERT_ID() 用于在同一连接中取得最近一次相关语句生成的自增值；多行插入时通常返回该语句生成的第一个自增值。不能用 MAX(id) 代替它获取当前连接刚插入的主键。

自动提交开启且没有显式事务时，成功的 INSERT 通常作为独立事务提交。显式事务中，尚未提交的插入可通过 ROLLBACK 撤销，但已分配的自增值通常不会回收，因此自增序列可能出现空缺。

## 6. InnoDB 加锁行为

普通 INSERT 涉及表级 IX 意向锁、插入意向锁和新记录的 X 排他保护；唯一键检查、外键检查与自增值分配还可能引入额外的锁和等待。

详细过程见[插入加锁流程](../innodb/locks/InsertLocking.md)。
