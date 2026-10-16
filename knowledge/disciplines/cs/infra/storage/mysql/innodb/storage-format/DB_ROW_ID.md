---
color: "#5B6CDA"
---

# DB_ROW_ID

## 1. 定义

DB_ROW_ID 是 InnoDB 在需要内部聚簇索引键时使用的隐藏行 ID，大小为 6 字节。它不是所有表的聚簇记录都具有的字段。

“按需存在”是对字段出现条件的说明，不属于系统列名称。

## 2. 何时使用

InnoDB 按以下顺序选择聚簇索引：

1. 使用显式主键。
2. 没有主键时，使用第一个所有列均为 NOT NULL 的唯一索引。
3. 没有合适的索引时，生成 DB_ROW_ID，建立内部聚簇索引 GEN_CLUST_INDEX。

```sql
CREATE TABLE events (
    message VARCHAR(200)
) ENGINE = InnoDB;
```

示例没有主键或唯一非空索引，因此 InnoDB 需要隐藏行 ID 组织记录。该表的二级索引若后来存在，也需要包含相应的内部聚簇索引键来定位行。

## 3. 源码标识

在常见 MySQL InnoDB 源码中，相应的系统列标识为 DATA_ROW_ID，长度常量为 DATA_ROW_ID_LEN，值为 6。

系统列标识存在于源码，不表示所有表的物理记录都保存 DB_ROW_ID；是否保存取决于聚簇索引的选择。

## 4. 与业务主键的区别

DB_ROW_ID 由 InnoDB 内部管理，不能直接通过 SELECT DB_ROW_ID 查询，也不能用作应用稳定访问记录的公开接口。

它不等于 AUTO_INCREMENT 列，不应依赖其值或分配顺序作为业务编号。通常应显式定义合适的主键，明确记录身份和索引组织方式。
