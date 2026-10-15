---
color: "#5B6CDA"
---

# 最左前缀

## 1. 定义与原理

对于常规 B+ 树联合索引，最左前缀是从索引定义的第一个列开始、连续使用前面的列进行定位的规则。“左”指索引列的定义顺序，不是 SQL 条件的书写顺序。

例如 `(name, age, city)` 先按 name 排序，再在相同 name 下按 age 排序，最后在相同 name、age 下按 city 排序。

age 只在相同 name 的记录中有序，并非在整个索引中全局有序。因此，仅指定 age 时，通常无法通过普通的前缀定位找到所有匹配记录所在的单一连续区间。

## 2. 查询示例

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    city VARCHAR(100) NOT NULL,
    INDEX idx_name_age_city (name, age, city)
) ENGINE = InnoDB;
```

| 查询条件 | 通常用于前缀定位的列 |
| --- | --- |
| `name = 'Alice'` | name |
| `name = 'Alice' AND age = 20` | name、age |
| `name = 'Alice' AND age = 20 AND city = 'Shanghai'` | name、age、city |
| `name = 'Alice' AND city = 'Shanghai'` | name；跳过 age 后，city 通常用于后续过滤。 |
| `age = 20` | 缺少最左列，通常无法进行普通的前缀定位。 |

SQL 条件顺序不会改变索引定义，因此以下查询也可以利用 name、age 前缀：

```sql
SELECT id FROM users WHERE age = 20 AND name = 'Alice';
```

## 3. 范围条件与后续列

```sql
SELECT id FROM users
WHERE name = 'Alice' AND age > 20 AND city = 'Shanghai';
```

通常可以用 name 和 age 确定扫描范围。由于不同 age 下的 city 分别排序，city 条件通常不能进一步把这个范围缩成一个连续区间。

后续列仍可能用于索引条件下推、过滤或覆盖查询。因此，“范围条件之后的列一律失效”并不准确；不同运算符对范围边界的构造也可能不同，应结合执行计划判断。

## 4. 其他访问方式

不满足最左前缀定位规则，不代表索引完全不能被使用。

- **全索引扫描**：扫描整个索引并过滤；如果索引覆盖查询，仍可能比扫描完整表记录更划算。
- **跳跃扫描**：特定版本和查询条件下，优化器可能枚举前导列的不同值，再使用后续列定位。MySQL 8.0 提供此优化，但有适用限制。

这些方式与直接使用最左前缀定位不同，是否采用由优化器根据成本决定。

## 5. 与前缀索引的区别

最左前缀讨论联合索引的列顺序；字符串前缀索引则只保存某列值的前若干字符，例如 `INDEX idx_name (name(20))`。两者不是同一个概念。

## 6. 查看执行计划

```sql
EXPLAIN SELECT id FROM users WHERE name = 'Alice' AND age = 20;
EXPLAIN SELECT id FROM users WHERE age = 20;
```

可以关注访问类型、所选索引、key_len 和 Extra。key_len 可辅助判断参与定位的索引键长度，但不能单独反映后续列是否用于过滤或覆盖查询。

[返回联合索引](Composite.md)
