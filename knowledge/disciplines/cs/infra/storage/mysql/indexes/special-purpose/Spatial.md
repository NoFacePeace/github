---
color: "#5B6CDA"
---

# 空间索引

## 1. 定义

空间索引通过 SPATIAL 定义，InnoDB 使用 R-tree 结构，利用空间对象的最小包围矩形筛选候选记录。精确空间关系通常仍需通过相应空间函数判断。

## 2. 示例

以下示例面向 MySQL 8.0，使用 SRID 0 的平面坐标：

```sql
CREATE TABLE places (
    id BIGINT PRIMARY KEY,
    location POINT NOT NULL SRID 0,
    SPATIAL INDEX sp_location (location)
) ENGINE = InnoDB;
```

## 3. 使用注意

InnoDB 空间索引列需要定义为 NOT NULL。MySQL 8.0 中，应为列显式指定 SRID，以便优化器考虑使用空间索引，并确保查询几何对象的 SRID 与列一致。

SRID 决定坐标参考系统，不能把平面坐标计算直接当作地理经纬度距离计算。实际是否使用索引需通过执行计划确认。

[返回特殊用途维度](SpecialPurpose.md)
