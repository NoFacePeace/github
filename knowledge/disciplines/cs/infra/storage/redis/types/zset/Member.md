---
color: "#5B6CDA"
---

# Member

member 是 ZSet 中的成员，用于标识一个元素。每个 member 关联一个 score；同一个 ZSet 内，member 唯一。

## 1. 与 Redis Key 的区别

```text
Redis Key：leaderboard
    ├─ member：alice，score：10
    └─ member：bob，score：20
```

`leaderboard` 定位整个集合，`alice` 和 `bob` 标识集合内部的元素。相同 member 可以存在于不同 ZSet 中。

## 2. 类型与相等判断

member 是二进制安全的字符串，可以保存文本、ID 或任意字节内容，不要求是数字或合法 UTF-8 文本。

是否为同一成员取决于完整字节内容：

- `alice` 与 `Alice` 不同，区分大小写。
- `1` 与 `01` 不同，不会因为能解释为相同数字而合并。
- Redis 不会对成员自动进行 Unicode 规范化；显示相同但字节不同的文本，也可能是不同成员。

## 3. 唯一性与更新

```redis
ZADD ranking 10 alice
ZADD ranking 20 alice
ZCARD ranking
ZSCORE ranking alice
```

最终只有一个 `alice`，分数为 `20`。第二次添加默认更新分数，而不是新增同名成员。

```redis
ZADD ranking NX 30 alice
```

`NX` 只允许新增不存在的成员，因此这次不会改变已有 `alice` 的分数。

## 4. 同分时的排序

ZSet 默认先按 score 升序排列，score 相同时按 member 的字节字典序排列，不按插入先后排序，也不按语言或地区的自然排序规则排序。

```redis
ZADD ranking 10 bob 10 alice
ZRANGE ranking 0 -1
```

返回顺序为 `alice`、`bob`。数字形式的成员也按字节比较，例如同分时字符串 `10` 排在 `2` 前面。

## 5. 业务使用

通常使用稳定的业务 ID 作为 member，将详情放在独立 KV 或数据库中。若把完整 JSON 作为 member，JSON 内容变化就会产生不同成员，无法自动按 JSON 内的某个 ID 去重。

member 不是独立的 Redis key，不能直接用 `GET member` 读取，也不能用 `EXPIRE member` 设置成员过期。`EXPIRE` 作用于整个 ZSet key；需要成员级过期时，可用 score 记录时间并自行清理。
