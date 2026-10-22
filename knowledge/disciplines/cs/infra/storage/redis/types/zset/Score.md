---
color: "#5B6CDA"
---

# Score

score 是 ZSet 中与 member 关联的数值，用于确定成员的排序位置。每个成员在同一时刻对应一个 score，不同成员的 score 可以相同。

## 1. 类型与范围

score 使用 IEEE 754 双精度浮点数表示，支持整数、小数、负数以及正负无穷，不接受 NaN。

```redis
ZADD ranking 10 alice -5 bob 1.5 carol
```

默认排序为 `bob`、`carol`、`alice`，对应分数为 `-5`、`1.5`、`10`。负分数和小数不需要额外的数据类型。

## 2. 排序与重复

```redis
ZADD ranking 20 bob 10 carol 20 alice
ZRANGE ranking 0 -1 WITHSCORES
```

逻辑顺序为：

```text
carol → 10
alice → 20
bob   → 20
```

score 可以重复，同分时按成员字节字典序排序。score 是排序依据，不是排名；排名从 `0` 开始，会随其他成员的增删或分数变化而改变。

## 3. 查询与修改

```redis
ZSCORE ranking alice
ZADD ranking 30 alice
ZINCRBY ranking 5 alice
ZINCRBY ranking -2 alice
```

`ZSCORE` 查询分数；`ZADD` 默认直接更新已有成员的分数；`ZINCRBY` 按增量调整分数，负增量表示减少。以上修改会自动维护成员的排序位置。

score 不是可存放任意 JSON 的业务值，适合积分、时间戳或优先级等数值。

## 4. 精度限制

双精度浮点数能够精确表示 `-2^53` 到 `2^53` 范围内的所有整数；超出这个范围，不保证相邻整数可被区分。部分更大的整数仍可精确表示，但不能据此假定任意 64 位整数都可无损用作 score。

需要特别注意：

- 不宜直接将任意 64 位 ID 当作精确排序分数。
- 当前量级的毫秒 Unix 时间戳在精确整数范围内，但纳秒时间戳通常超出范围。
- 将时间戳乘以较大倍数再拼接序号，可能超出精度范围。
- 小数采用二进制浮点表示，不适合直接承担要求严格十进制精度的金额计算；可根据范围采用整数最小单位。

## 5. 用于时间与优先级

```redis
ZADD pending 1792630800000 session-a
ZADD pending 1792630801000 session-b
ZRANGE pending 0 -1
```

用入队时间作 score 时，较早入队的成员先被读取；同一时间戳仍按 member 排序。用到期时间作 score 时，可以筛选已经到期的成员，但时间到达不会自动删除成员，需要消费者或清理程序执行删除。

用优先级作 score 时，需要约定较小数字还是较大数字代表高优先级，并选择对应的升序或降序读取方式。
