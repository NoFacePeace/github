---
color: "#5B6CDA"
---

# SELECT

SELECT 用于查询数据，可以选择列、筛选记录、连接多张表、分组聚合，以及控制结果顺序和数量。它也可以计算表达式，不一定需要 FROM，例如 `SELECT 1 + 1;`。

## 1. 基本语法

常见查询结构如下，方括号表示可选部分，并非完整语法：

```sql
SELECT [DISTINCT] select_list
FROM table_reference
[WHERE condition]
[GROUP BY column_list]
[HAVING group_condition]
[ORDER BY sort_list]
[LIMIT row_count OFFSET offset];
```

以下示例假设 users 表包含 id、name、age 和 department_id 列，且 id 是主键：

```sql
SELECT id, name AS user_name
FROM users
WHERE age >= 18
ORDER BY id
LIMIT 10;
```

SELECT 指定返回列或表达式，AS 设置别名。`SELECT *` 返回所有列，业务查询通常显式列出需要的列，以减少数据传输并明确结果结构。

## 2. 条件过滤与去重

WHERE 在分组前筛选记录，常用条件包括比较运算、AND、OR、IN、BETWEEN、LIKE 和 IS NULL。

```sql
SELECT DISTINCT department_id
FROM users
WHERE age BETWEEN 18 AND 30
  AND department_id IS NOT NULL;
```

BETWEEN 包含两端边界。DISTINCT 对所选列组成的整行结果去重，而不是只对第一列去重。

判断空值应使用 IS NULL 或 IS NOT NULL；`column = NULL` 不会得到 TRUE。WHERE 只保留条件为 TRUE 的记录，条件为 FALSE 或 UNKNOWN 的记录都会被过滤。

## 3. 多表连接

以下示例额外假设 departments 表包含 id 和 name 列，且 id 是主键：

```sql
SELECT u.id, u.name, d.name AS department_name
FROM users AS u
LEFT JOIN departments AS d
  ON d.id = u.department_id
WHERE u.age >= 18;
```

INNER JOIN 只返回满足连接条件的组合；LEFT JOIN 保留左表记录，没有匹配的右表列以 NULL 补齐。一条记录匹配右表多条记录时，结果也会包含多行。

LEFT JOIN 中，ON 决定右表匹配范围，WHERE 再筛选连接结果。如果在 WHERE 中加入 `d.name = 'Engineering'`，未匹配右表的记录会被过滤；若要保留这些左表记录，应根据业务语义考虑将该条件放在 ON 中。

## 4. 分组与聚合

GROUP BY 将记录分组，COUNT、SUM、AVG、MIN、MAX 等函数计算聚合结果，HAVING 筛选分组后的结果：

```sql
SELECT department_id, COUNT(*) AS user_count, AVG(age) AS average_age
FROM users
WHERE age >= 18
GROUP BY department_id
HAVING COUNT(*) >= 5
ORDER BY user_count DESC, department_id;
```

COUNT(*) 统计行数，COUNT(column) 统计该列非 NULL 的行数；SUM 和 AVG 等聚合函数通常忽略 NULL。

启用 ONLY_FULL_GROUP_BY 时，非聚合表达式通常需要出现在 GROUP BY 中，或能由分组列通过函数依赖确定。不要依赖关闭该模式后任意选取组内值的行为。

## 5. 排序与分页

ORDER BY 明确结果顺序，ASC 表示升序，DESC 表示降序。没有 ORDER BY 时，不应依赖返回顺序；排序字段相同时，可增加唯一列作为最终排序键，以获得稳定顺序。

```sql
SELECT id, name
FROM users
ORDER BY id
LIMIT 10 OFFSET 20;
```

此查询跳过前 20 行，最多返回 10 行。较大的 OFFSET 可能需要扫描并跳过大量记录。按唯一递增 id 翻页时，可以使用上一页最后一个 id 继续查询：

```sql
SELECT id, name
FROM users
WHERE id > 1000
ORDER BY id
LIMIT 10;
```

实际应用需传入上一页最后一个 id，并保持相同的筛选和排序条件。分页期间并发修改数据仍可能影响跨页结果；复杂排序还需要使用对应的排序键构造游标条件。

## 6. 逻辑处理与执行计划

理解常见查询语义时，可以按 FROM / JOIN → WHERE → GROUP BY → HAVING → SELECT → DISTINCT → ORDER BY → LIMIT 的顺序分析。这是概念上的逻辑顺序，不代表引擎逐步执行的物理过程。

MySQL 允许在 GROUP BY、HAVING 和 ORDER BY 中引用选择列表的别名，但 WHERE 通常不能使用同层 SELECT 定义的别名。

实际执行时，优化器会选择访问路径、连接顺序等策略。使用 EXPLAIN 查看计划；需要实际执行统计时，可以结合版本使用 EXPLAIN ANALYZE。

## 7. 普通查询与锁定读取

InnoDB 中，普通 SELECT 通常使用一致性读，通过事务隔离级别和 Read View 决定可见版本。`SELECT ... FOR UPDATE` 和 `SELECT ... FOR SHARE` 属于锁定读取，读取并锁定相应记录，具体锁范围与索引、条件及隔离级别有关。

锁定读取应放在明确的事务中使用，相关锁通常在事务提交或回滚时释放。不能把所有 SELECT 都理解成无锁读取，普通查询也可能涉及元数据锁。
