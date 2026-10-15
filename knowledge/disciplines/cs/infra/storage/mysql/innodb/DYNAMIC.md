---
color: "#5B6CDA"
---

# DYNAMIC

## 1. 定义与默认值

DYNAMIC 是 InnoDB 的一种行格式，规定记录内部的编码方式及长字段的页外存储策略。MySQL 5.7.9 起，InnoDB 默认行格式为 DYNAMIC，但实例可以通过 innodb_default_row_format 修改默认值。

默认配置不代表所有已有表都采用 DYNAMIC；旧表可能保留其他格式，新表也可以显式指定受支持的其他格式。名称中的“动态”不表示查询时自动切换行格式。

## 2. 页内与页外存储

对于适用的长变长字段，需要页外存储时，DYNAMIC 可以将字段值整体移到页外，聚簇索引叶子记录中仅保留 20 字节引用。

```text
聚簇索引叶子记录
├─ id：10
├─ title：文章标题
└─ content：20 字节引用
                 ↓
             页外数据
```

这可以减少大字段对索引叶子页的占用，让页内保留更多记录。但读取页外字段时，也可能需要访问额外的数据页。

是否放到页外取决于字段类型、实际长度、整条记录大小和页大小等因素。VARCHAR、TEXT、BLOB 不一定都在页外，短字段可以保存在行内。外置字段也不会让行和列的大小限制消失。

## 3. 与 COMPACT 的区别

| 对比项 | DYNAMIC | COMPACT |
| --- | --- | --- |
| 需要页外存储的适用长字段 | 可以完整放到页外，行内保留 20 字节引用。 | 通常在行内保留前 768 字节和 20 字节引用。 |
| 长字段的行内占用 | 可以较小。 | 保留的前缀可能占用更多页内空间。 |
| 基础记录编码 | 沿用 COMPACT 风格的基础结构，包括记录头、NULL 位图及变长字段长度信息等。 | 使用相应的紧凑编码结构。 |

这里的差异描述长字段实际采用页外存储时的情况，不表示每个字段都固定采用这种布局。

## 4. 不同索引记录的内容

同一表的常规索引记录遵循相应的格式编码规则，但不同节点的字段与布局并不完全相同：

- [聚簇索引叶子记录](ClusteredIndexLeafRecord.md)：包含用户列和必要的事务内部信息，适用的长字段可通过引用存到页外。
- **二级索引叶子记录**：包含索引列与聚簇索引键，并非完整业务行。
- **非叶子记录**：包含导航键与子页编号，并非业务行，也不包含聚簇叶子记录那套完整的事务隐藏字段。

DYNAMIC 的页外存储优势主要针对聚簇索引中的长列值，不能据此认为二级索引键也可以无限长；索引键长度仍受相应限制。

## 5. 指定与查看

```sql
CREATE TABLE articles (
    id BIGINT PRIMARY KEY,
    title VARCHAR(200),
    content TEXT
) ENGINE = InnoDB ROW_FORMAT = DYNAMIC;

SHOW VARIABLES LIKE 'innodb_default_row_format';
SHOW TABLE STATUS LIKE 'articles';
SHOW CREATE TABLE articles;
```

SHOW TABLE STATUS 的 Row_format 用于查看实际格式；默认配置只影响未显式指定格式的新建等操作，不会自动重写已有表。

已有表可以使用 `ALTER TABLE articles ROW_FORMAT = DYNAMIC` 调整格式，通常涉及表重建，应结合版本、表规模和在线 DDL 支持情况评估。

[返回行格式](RowFormat.md)
