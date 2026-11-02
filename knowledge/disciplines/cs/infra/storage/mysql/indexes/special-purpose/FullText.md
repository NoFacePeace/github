---
color: "#5B6CDA"
---

# 全文索引

## 1. 定义

全文索引通过 `FULLTEXT` 定义，用于根据文本中的词项查找记录，并计算相关性分数。MySQL 8.0 的 InnoDB 和 MyISAM 支持全文索引，MEMORY 不支持；适用列类型为 `CHAR`、`VARCHAR` 和 `TEXT`。

全文检索通常包含文本解析、倒排索引、词项匹配和相关性评分。倒排索引记录“词项 → 文档”的映射，与普通 B+ 树索引的组织方式不同。

全文索引的维护、匹配与评分依赖存储引擎的实现，不同引擎的行为可能不同。下面以 MySQL 8.0 的 InnoDB 为例。

## 2. 默认解析器与基本用法

创建全文索引时不指定 `WITH PARSER`，就使用默认解析器：

```sql
CREATE TABLE articles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    FULLTEXT INDEX ft_title_body (title, body)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

INSERT INTO articles (title, body) VALUES
('MySQL search', 'MySQL supports full text search'),
('Redis caching', 'Redis supports caching');

SELECT id, title FROM articles
WHERE MATCH(title, body) AGAINST('search' IN NATURAL LANGUAGE MODE);
```

这条查询匹配第一条记录。建立和维护索引时，MySQL 自动解析文本；查询时也会解析搜索文本，应用无需预先手动分词。

默认解析器主要根据空格、标点等边界划分单词，适合英文等文本，不会把连续中文按自然语言词语切开。InnoDB 默认最小索引词长为 3，最大为 84，由 `innodb_ft_min_token_size` 和 `innodb_ft_max_token_size` 控制；停用词也可能被忽略。

## 3. 中文与 ngram 解析器

中文等文本可以使用内置的 `ngram` 解析器，在创建索引时指定：

```sql
CREATE TABLE chinese_articles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    FULLTEXT INDEX ft_title_body (title, body) WITH PARSER ngram
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

INSERT INTO chinese_articles (title, body) VALUES
('MySQL 索引', '通过索引提升数据库查询性能'),
('MySQL 事务', '事务支持提交和回滚');

SELECT id, title FROM chinese_articles
WHERE MATCH(title, body) AGAINST('数据库' IN NATURAL LANGUAGE MODE);
```

`ngram` 按固定长度的连续字符切分，默认 `ngram_token_size = 2`。例如“数据库查询”切分为“数据”“据库”“库查”“查询”，并不理解“数据库”是一个完整词语。

`ngram` 使用自己的 token 长度配置，不受 InnoDB 默认解析器的最小、最大词长配置控制。修改词长、ngram 大小或停用词等索引构建配置后，需要按对应配置要求重建已有全文索引。

## 4. 查询模式

| 模式 | 用法与含义 |
| --- | --- |
| 自然语言模式 | `IN NATURAL LANGUAGE MODE`，根据查询词项匹配并计算相关性；省略模式时默认使用此模式。 |
| 布尔模式 | `IN BOOLEAN MODE`，支持 `+`、`-` 等操作符，表达必须包含、必须排除等条件。 |
| 查询扩展 | `WITH QUERY EXPANSION`，利用第一轮匹配结果中的词项扩展搜索，可能增加召回，也可能引入不相关结果。 |

例如，在使用默认解析器的 `articles` 表中：

```sql
SELECT id, title FROM articles
WHERE MATCH(title, body) AGAINST('+MySQL -caching' IN BOOLEAN MODE);
```

`+MySQL` 要求匹配该词，`-caching` 排除匹配该词的记录。实际匹配单位仍由解析器决定；使用 ngram 时不能简单把每段查询文本当作一个完整词。

## 5. 相关性评分与 Top K

`MATCH ... AGAINST` 可以返回相关性分数，按分数降序排列并使用 `LIMIT` 即可返回 Top K：

```sql
SELECT
    id,
    title,
    MATCH(title, body)
        AGAINST('search' IN NATURAL LANGUAGE MODE) AS score
FROM articles
WHERE MATCH(title, body)
        AGAINST('search' IN NATURAL LANGUAGE MODE)
ORDER BY score DESC, id ASC
LIMIT 10;
```

分数越高表示对当前查询越相关；`id ASC` 用于稳定同分记录的顺序。建议显式指定 `ORDER BY`，尤其是布尔模式不会自动按相关性降序返回结果。`LIMIT 10` 限制返回数量，不代表数据库只需检查 10 条候选。

InnoDB 的全文检索评分参考 BM25 和 TF-IDF，但不能直接等同于标准 BM25。MySQL 8.0 原生全文检索不提供切换 BM25、TF-IDF 等评分算法的选项，也不提供标准 BM25 的 `k1`、`b` 调参接口。选择解析器或查询模式，不等于选择评分算法。

## 6. 使用注意

全文检索不是任意子串匹配，不能直接等同于 `LIKE '%关键词%'`。搜索结果受分词器、停用词、词长配置和查询模式影响。中文等文本需要评估 ngram 分词器等适合的配置。

对于 InnoDB，`MATCH` 中的列列表应与某个全文索引的列列表对应。例如，已有 `(title, body)` 全文索引时，若需要单独执行 `MATCH(title)`，应另建 `FULLTEXT(title)` 索引。

InnoDB 全文索引只处理已提交的数据，事务中的新增文本需要提交后才能被全文检索查到。

全文检索依据词项匹配，不提供向量语义相似度搜索。MySQL 8.0 可以通过 `JSON` 或 `BLOB` 保存向量，但没有原生向量类型和向量索引，不能把全文检索的 Top K 当作向量最近邻检索。
