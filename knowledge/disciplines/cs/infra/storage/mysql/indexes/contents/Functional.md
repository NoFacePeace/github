---
color: "#5B6CDA"
---

# 函数索引

## 1. 定义

函数索引保存表达式的计算结果，MySQL 8.0.13 起支持函数索引键部分，内部通过隐藏的虚拟生成列实现。

## 2. 示例

```sql
CREATE INDEX idx_lower_name ON users ((LOWER(name)));

EXPLAIN SELECT * FROM users WHERE LOWER(name) = 'alice';
```

示例假设 `users` 已包含 `name` 列。定义中的表达式需要额外的一层括号。

## 3. 使用注意

查询表达式及其结果类型需要与索引表达式相匹配，优化器才可能使用该索引。表达式受到生成列相关规则限制，不能任意使用非确定性函数、子查询或存储函数。

大小写比较行为也受字符集和排序规则影响，不应仅因查询涉及大小写转换就增加函数索引。

[返回索引内容维度](Contents.md)
