---
color: "#5B6CDA"
---

# 全文索引

## 1. 定义

全文索引通过 FULLTEXT 定义，用于文本搜索，常见支持引擎包括 InnoDB 和 MyISAM。适用列类型为 CHAR、VARCHAR 和 TEXT。

## 2. 示例

```sql
CREATE TABLE articles (
    id BIGINT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    FULLTEXT INDEX ft_title_body (title, body)
) ENGINE = InnoDB;

SELECT id, title FROM articles
WHERE MATCH(title, body) AGAINST('database' IN NATURAL LANGUAGE MODE);
```

## 3. 使用注意

全文检索不是任意子串匹配，不能直接等同于 `LIKE '%关键词%'`。搜索结果受分词器、停用词、词长配置和查询模式影响。中文等文本需要评估 ngram 分词器等适合的配置。

[返回特殊用途维度](SpecialPurpose.md)
