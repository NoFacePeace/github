---
color: "#5B6CDA"
---

# MySQL 索引合并

## 1. 定义

索引合并（Index Merge）是对同一张表进行多次索引范围扫描，再对匹配记录的行标识求交集、并集或组合处理的访问方式。通常涉及多个索引，而不是将多个索引结构合成一个联合索引。

它不用于合并不同表的索引。InnoDB 通过二级索引取得主键后访问聚簇索引的回表过程，也不属于索引合并。

## 2. 示例表

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    city VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    INDEX idx_city (city),
    INDEX idx_age (age)
) ENGINE = InnoDB;
```

以下 SQL 展示可能的候选路径，不保证优化器一定采用索引合并。

## 3. 交集

```sql
SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;
```

可以分别扫描 idx_city 和 idx_age，只保留两套索引共同匹配的记录，再按需要读取完整行。执行计划可能显示 `Using intersect(idx_city,idx_age)`。

交集合并对访问方式有要求。常见适用条件包括覆盖一个索引全部键部分的等值条件，以及 InnoDB 主键上的范围条件。不能认为任意两个范围条件都能使用交集合并。

在 InnoDB 中，二级索引全键等值匹配后的记录按主键排列，这类顺序有助于按行标识合并。若合并使用的索引已经提供查询所需的全部列，可能不必为获取列值读取完整行；否则需要访问聚簇索引。

## 4. 并集与排序并集

### 4.1 并集

```sql
SELECT * FROM users WHERE city = 'Shanghai' OR age = 20;
```

分别取得两个条件的匹配记录，再合并去重。同一记录满足两个条件时，结果不会因此重复。访问方式满足要求时，执行计划可能显示 `Using union(idx_city,idx_age)`。

### 4.2 排序并集

```sql
SELECT * FROM users WHERE city < 'M' OR age > 60;
```

当可进行多索引范围扫描，但不满足直接并集合并的条件时，可能使用 sort_union。它先收集行标识，再排序和去重，因此带来额外成本，执行计划可能显示 `Using sort_union(idx_city,idx_age)`。

AND、OR 组合的查询也可能形成交集的并集等计划，具体取决于条件形式和优化器支持。

## 5. 何时选择索引合并

优化器需要同时判断可用性和预计成本，包括：

- 各索引需要扫描多少条记录。
- 合并、排序及去重需要多少开销。
- 合并后需要读取多少完整行。
- 单索引加过滤、联合索引或全表扫描是否更便宜。

即使最终交集很小，两套索引各自扫描大量记录也可能很昂贵。AND、OR 只是查询语义，不是使用索引合并的保证。全文索引不支持这种 Index Merge 访问方式。

## 6. 与联合索引的选择

经常使用 `city = ... AND age = ...` 时，可以考虑联合索引 `(city, age)`，直接按组合条件定位，通常避免扫描两套索引再求交集。但是否值得建立，还要结合其他查询与写入成本判断。

对于 `city = ... OR age = ...`，联合索引 `(city, age)` 通常不能通过一次最左前缀定位覆盖两边的条件，因此不能简单替代两个单列索引。

## 7. 查看执行计划

```sql
EXPLAIN SELECT * FROM users WHERE city = 'Shanghai' AND age = 20;
EXPLAIN SELECT * FROM users WHERE city = 'Shanghai' OR age = 20;
```

传统 EXPLAIN 中可关注：

| 字段 | 典型信息 |
| --- | --- |
| type | index_merge |
| key | 参与扫描的多个索引 |
| key_len | 各索引使用的键长度 |
| Extra | Using intersect、Using union 或 Using sort_union |

计划未出现 index_merge 并不代表异常，应比较实际成本。MySQL 8.0.18 起可用 EXPLAIN ANALYZE 检查实际行数与耗时，但它会执行查询。需要了解候选方案被放弃的原因时，可以查看 optimizer_trace。
