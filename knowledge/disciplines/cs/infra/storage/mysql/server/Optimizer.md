---
color: "#5B6CDA"
---

# MySQL 优化器

优化器位于 Server 层，根据查询语义、统计信息和成本模型选择[执行计划](ExecutionPlan.md)。索引选择只是其中一部分，还需要考虑连接顺序、排序、分组等操作。

- [索引选择](IndexSelection.md)：候选访问路径、统计信息、成本比较、索引合并与计划诊断。

## 1. 选择过程

1. 分析并进行适用的查询变换，找出满足语义要求的候选访问路径。
2. 估算各路径需要扫描的记录数量，以及过滤后保留的记录数量。
3. 估算扫描、回表、条件计算、连接和排序等成本。
4. 在考虑的候选方案中选择预计成本较低的计划，交给执行器。

优化器不会把所有候选计划实际执行一遍，也不保证找到理论上的全局最优方案。

## 2. 候选访问路径

| 路径 | 说明 |
| --- | --- |
| 索引等值或范围访问 | 根据条件定位部分索引记录。 |
| 全索引扫描 | 扫描整个索引，例如利用覆盖能力减少读取完整行的成本。 |
| 索引合并 | 对同一张表使用多个索引，并合并匹配记录。 |
| 全表扫描 | 扫描表记录；InnoDB 通常通过扫描聚簇索引完成。 |

联合索引的列顺序、条件表达式、类型和排序规则等会影响可用路径。一个索引可以覆盖查询，但不一定能够高效定位匹配记录。

## 3. 行数估算与统计信息

优化器利用表与索引统计信息、适用的索引探测和直方图等估算选择性及匹配行数。索引基数反映不同键值的数量，但不等于某个具体值一定匹配多少行。

统计信息通常带有采样和估算误差。数据分布不均、列之间存在相关性或统计信息过旧，都可能造成估计行数与实际行数偏差。

## 4. 成本比较示例

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    city VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    INDEX idx_city (city),
    INDEX idx_age (age)
) ENGINE = InnoDB;

SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;
```

| 候选方案 | 主要成本 |
| --- | --- |
| idx_city | 扫描匹配城市的索引记录，回表并过滤年龄。 |
| idx_age | 扫描匹配年龄的索引记录，回表并过滤城市。 |
| 索引合并交集 | 扫描两套索引、求交集，再按需要回表。 |
| 全表扫描 | 扫描全部表记录并判断条件。 |

若估计上海用户有 100 万人、20 岁用户只有 1 万人，其他条件相近时通常更倾向于 idx_age。若存在 `(city, age)` 联合索引，则可能直接按组合条件定位。

扫描行数不是唯一标准。覆盖索引可以节省回表；支持 ORDER BY 的索引可能减少排序；配合 LIMIT 提前停止读取，也可能改变方案成本。

## 5. 索引合并

同一张表的多个条件分别适用不同索引，且合并结果的预计成本较低时，优化器可能选择 Index Merge。

- **并集**：例如 `city = 'Shanghai' OR age = 20`，分别获取匹配记录后合并去重，可能使用 union 或 sort_union 算法。
- **交集**：例如 `city = 'Shanghai' AND age = 20`，在访问方式符合要求时，可以保留两套索引共同匹配的记录。

并非所有 AND、OR 条件都能或都会使用索引合并。优化器还会比较单个索引加过滤、联合索引和全表扫描等路径。

通常每张表的一次访问使用一个索引或全表扫描，索引合并是使用多个索引的例外。多表查询和同一表的多次引用可以分别选择访问路径。二级索引查找后回表访问聚簇索引，不称为索引合并。

## 6. 查看执行计划

```sql
EXPLAIN SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;

EXPLAIN FORMAT=JSON
SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;
```

常用检查项包括：

- **type**：访问方式；index_merge 表示索引合并，index 通常表示全索引扫描，ALL 表示全表扫描。
- **possible_keys 与 key**：可能用于定位的候选索引及实际选择的索引，两者不一定相同。
- **rows 与 filtered**：预计扫描行数与过滤后的保留比例，不是实际执行结果。
- **Extra**：覆盖索引、索引条件下推、额外排序等信息；索引合并可能显示 Using intersect、Using union 或 Using sort_union。

MySQL 8.0.18 起可以使用 EXPLAIN ANALYZE 对比预计与实际行数、耗时。它会实际执行查询，应考虑查询成本。

## 7. 诊断选择偏差

先检查实际数据分布、估计与实际行数的偏差，以及查询是否适合现有索引。

```sql
-- 更新表和索引统计信息，会读取数据并可能影响在线负载。
ANALYZE TABLE users;

-- 在当前会话中跟踪优化过程。
SET optimizer_trace = 'enabled=on';
EXPLAIN SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;
SELECT TRACE FROM INFORMATION_SCHEMA.OPTIMIZER_TRACE;
SET optimizer_trace = 'enabled=off';
```

optimizer_trace 可以帮助查看候选方案及成本比较。必要时可评估直方图、调整索引或查询，再验证实际效果。索引提示适合在理解计划偏差后谨慎使用，不能替代统计信息和长期维护。

[返回 Server 层](Server.md)
