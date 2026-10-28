---
color: "#5B6CDA"
---

# MKSTREAM

`MKSTREAM` 是 `XGROUP CREATE` 的可选选项，表示目标 Stream 不存在时，先创建空 Stream，再创建消费者组。`MK` 是 make（创建）的缩写，`MKSTREAM` 即 make stream（创建流）。

## 1. 用法

```redis
XGROUP CREATE orders workers 0 MKSTREAM
```

`CREATE` 的主要操作是创建消费者组 `workers`；`MKSTREAM` 附带启用按需创建流 `orders` 的行为，不需要额外的选项值。

## 2. 对目标 Stream 的处理

| 目标状态 | 带 `MKSTREAM` | 不带 `MKSTREAM` |
| --- | --- | --- |
| key 不存在 | 创建空 Stream，再创建组。 | 报错，不创建组。 |
| key 已存在且为 Stream | 使用已有 Stream 创建组，保留已有消息。 | 使用已有 Stream 创建组。 |
| key 已存在但类型不是 Stream | 返回类型错误。 | 返回类型错误。 |

`MKSTREAM` 不会覆盖已有 key，也不会忽略同名消费者组已存在时的 `BUSYGROUP` 错误。

## 3. 适用范围

`MKSTREAM` 专用于 `XGROUP CREATE`，其他 `XGROUP` 子命令不支持该选项。它不是独立命令。

`XADD` 在目标 key 不存在时，默认就会创建 Stream 并追加消息，因此无需使用 `MKSTREAM`。
