---
color: "#5B6CDA"
---

# InnoDB 行格式

## 1. 定义

行格式（Row Format）决定 InnoDB 如何将字段和管理信息编码成存储记录，包括字段排列、NULL 标记、变长字段长度，以及长字段的页内与页外存储方式。

聚簇索引决定行按什么键组织，行格式决定每条记录内部如何保存。行格式也影响二级索引记录的编码，但二级索引不保存完整行和聚簇记录中的整套事务隐藏列。

## 2. 记录内部结构

以 COMPACT、DYNAMIC 等格式的聚簇索引叶子记录为例，主要包含：

- **记录头**：保存记录类型、删除标记和记录链接等管理信息。
- **NULL 位图**：标记允许为 NULL 的列是否为空；NULL 与空字符串不是同一含义。
- **变长字段长度信息**：帮助确定 VARCHAR 等字段的实际存储长度。
- **字段数据**：保存用户列值，长字段可能包含指向页外数据的引用。
- **事务内部信息**：例如 DB_TRX_ID 和 DB_ROLL_PTR；缺少合适聚簇索引键时还有隐藏的 DB_ROW_ID。

这些内容的具体顺序与编码规则取决于格式和记录类型，不能把上述列表当作所有记录统一的字节布局。

## 3. 常见行格式

| 格式 | 特点 |
| --- | --- |
| [DYNAMIC](DYNAMIC.md) | 现代 MySQL 的默认格式；需要页外存储时，长变长字段可以整体移到页外，行内保留 20 字节引用。 |
| COMPACT | 较早的常用格式；适用的长字段移到页外时，通常在行内保留前 768 字节及 20 字节引用。 |
| REDUNDANT | 旧格式，主要用于兼容，其记录元数据编码方式不同。 |
| COMPRESSED | 支持压缩页，页外存储方式与 DYNAMIC 类似，有页大小、表空间和版本等限制。 |

这里的默认格式可通过 innodb_default_row_format 配置查看。COMPRESSED 行格式与透明页压缩不是同一个机制。

## 4. 页内与页外存储

InnoDB 以页管理记录。一行较大、难以满足页内存储要求时，可以将适用的长变长字段放到页外，由行内引用定位。

```sql
CREATE TABLE articles (
    id BIGINT PRIMARY KEY,
    title VARCHAR(200),
    content TEXT
) ENGINE = InnoDB ROW_FORMAT = DYNAMIC;
```

content 较短时可能保存在行内；需要页外存储时，DYNAMIC 可以将其整体移出，行内保留引用。不能把 TEXT、BLOB 或 VARCHAR 类型直接等同于“必定页外存储”。

是否移到页外取决于字段类型、实际长度、整条记录的大小和页大小等因素。页外存储也不代表行可以无限大，列数、行内记录大小和字段类型等限制仍然存在。

## 5. 查看与指定

```sql
-- 查看已有表的实际行格式。
SHOW TABLE STATUS LIKE 'articles';

-- 查看建表定义及显式选项。
SHOW CREATE TABLE articles;

-- 查看新建表使用的默认行格式配置。
SHOW VARIABLES LIKE 'innodb_default_row_format';
```

SHOW TABLE STATUS 的 Row_format 字段可用于查看实际格式。SHOW CREATE TABLE 不一定显式列出所有使用默认值的选项。

创建表时可以使用 ROW_FORMAT 指定格式；修改已有表也可使用：

```sql
ALTER TABLE articles ROW_FORMAT = DYNAMIC;
```

修改行格式通常涉及表重建，会消耗时间和存储资源，在线执行能力取决于版本和表的具体条件。仅修改默认行格式配置，不会自动重写已有表。
