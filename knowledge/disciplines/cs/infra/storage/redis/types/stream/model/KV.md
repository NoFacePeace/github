---
color: "#5B6CDA"
---

# KV

Stream 消息的内容是一组有序的 field-value 对，也可以称为键值对（KV）。这里的 key 指消息中的字段名，不是 Redis 数据库中的 key；使用 field-value 表述更明确。

## 1. 字段与值

```redis
XADD orders * order_id 1001 status created
```

| 部分 | 含义 |
| --- | --- |
| `orders` | Redis 中的 Stream key。 |
| `*` | 自动生成消息 ID。 |
| `order_id 1001` | 第一个 field-value 对。 |
| `status created` | 第二个 field-value 对。 |

字段和值都是二进制安全的字符串。示例中的 `1001` 作为字符串保存，不是具有独立整数类型的值。一个字段值可以存放 JSON 等序列化内容，其解析由应用负责。

## 2. 字段结构

每条消息至少包含一个 field-value 对。不同消息可以使用不同的字段，不要求整条 Stream 具有统一的字段结构。

同一条消息中的字段按写入顺序保存，允许重复字段名。例如：

```redis
XADD events * tag payment tag retry
```

这条消息包含两个名为 `tag` 的字段。客户端如果将消息转换为普通字典或映射，可能丢失重复字段或顺序信息，需要留意客户端的返回形式。

## 3. 与 Hash 的区别

消息内容不是独立的 Redis Hash，不能通过 `HGET`、`HSET` 操作。Hash 中字段名唯一，而 Stream 消息内容是允许重复字段名的有序字段值对序列。

消息 ID 标识整条消息，field-value 对描述消息内容；业务字段 `order_id` 等不会替代消息 ID，也不受消息 ID 的递增规则约束。
