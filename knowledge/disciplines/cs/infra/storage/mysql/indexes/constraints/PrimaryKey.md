---
color: "#5B6CDA"
---

# 主键索引

## 1. 定义

主键用于唯一标识一行，键值唯一且不允许 NULL。一张表只能定义一个主键，但主键可以由多个列组成。

## 2. 示例

```sql
CREATE TABLE users (
    id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id)
) ENGINE = InnoDB;
```

## 3. 存储引擎关系

InnoDB 使用显式主键组织聚簇索引，叶子节点存储完整行，二级索引记录中也包含主键值。因此通常应选择短小、稳定的主键；这种存储方式不能推广到所有引擎。
