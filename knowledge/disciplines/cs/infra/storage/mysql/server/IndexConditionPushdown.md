---
color: "#5B6CDA"
---

# 索引条件下推

## 1. 定义

索引条件下推（Index Condition Pushdown，ICP）将能够仅凭索引记录判断的部分过滤条件交给存储引擎，让引擎在读取完整行之前检查，减少不必要的行读取。在 InnoDB 二级索引访问中，主要体现为减少回表。

传递的是内部条件表达式，不是向存储引擎重新发送一条 SQL。不能下推的剩余条件仍由 Server 层处理。

## 2. 各层分工

| 组件 | 职责 |
| --- | --- |
| [优化器](Optimizer.md) | 结合访问路径，判断哪些条件适合下推，形成执行安排。 |
| [执行器](Executor.md)及访问接口 | 按计划驱动读取，将适用条件交给存储引擎，并处理剩余条件。 |
| [存储引擎](../StorageEngines.md) | 扫描索引记录，检查下推条件，符合时才继续读取完整行。 |

ICP 是 Server 与存储引擎协作的优化，不是 InnoDB 独有机制；MyISAM 等引擎也可以支持。

## 3. 查询示例

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    email VARCHAR(200) NOT NULL,
    INDEX idx_name_age (name, age)
) ENGINE = InnoDB;

SELECT * FROM users
WHERE name LIKE '张%' AND age = 20;
```

假设优化器选择 idx_name_age 的范围访问，可以区分两项工作：

1. **确定扫描范围**：利用 name 的前缀匹配定位候选索引区间。
2. **检查下推条件**：在候选索引记录上检查 age 是否等于 20，符合才回表读取 email 等所需列。

由于 name 是范围条件，age 通常不能继续缩小为单一连续扫描区间，但它仍存在于索引记录中，可以用于过滤。

## 4. 使用前后的流程

| 未使用 ICP | 使用 ICP |
| --- | --- |
| 扫描姓名符合范围的索引记录。 | 扫描相同范围的索引记录。 |
| 对候选记录回表读取完整行。 | 先在索引记录上判断 age = 20。 |
| 将记录交给 Server 层检查年龄。 | 只有年龄符合的记录才回表并返回。 |

若索引范围内有 1 万条记录，只有 100 条满足年龄条件，ICP 可能将为获取完整行而进行的回表从约 1 万次减少到约 100 次。

它减少的是完整行读取，不一定减少扫描的索引记录数量。实际收益还受缓存命中率和条件选择性等因素影响。

## 5. 适用范围与限制

ICP 常用于 range、ref、eq_ref 和 ref_or_null 等访问方式，且查询需要读取完整行的情况。具体支持范围需结合 MySQL 版本和存储引擎判断。

在 InnoDB 中，ICP 用于二级索引，不用于聚簇索引：聚簇索引叶子记录本身已包含行数据，无法获得同样的减少回表收益。

下推条件必须适合通过索引记录计算，但并非所有引用索引列的表达式都能下推。例如，涉及子查询或存储函数的条件不支持 ICP，InnoDB 虚拟生成列上的二级索引也不支持 ICP。

## 6. 与其他索引机制的区别

| 机制 | 解决的问题 |
| --- | --- |
| 索引范围定位 | 缩小需要扫描的索引区间。 |
| 索引条件下推 | 在候选索引记录上提前过滤，减少完整行读取。 |
| 覆盖索引 | 索引已包含查询所需列，通常无需为了获取列值回表。 |

不能根据“联合索引后续列无法继续缩小扫描范围”，就认定这些列完全没有作用。它们仍可能用于 ICP、其他过滤或覆盖查询。

## 7. 查看执行计划与配置

```sql
EXPLAIN SELECT * FROM users
WHERE name LIKE '张%' AND age = 20;

SHOW VARIABLES LIKE 'optimizer_switch';
```

传统 EXPLAIN 的 Extra 出现 `Using index condition`，表示使用索引条件下推。`Using index` 通常表示覆盖索引，与 ICP 含义不同。

ICP 由 optimizer_switch 中的 index_condition_pushdown 开关控制，通常默认开启。是否采用仍由访问路径、条件适用性等因素决定，不能因为开启开关就认定某条查询一定使用 ICP。

[返回 Server 层](Server.md)
