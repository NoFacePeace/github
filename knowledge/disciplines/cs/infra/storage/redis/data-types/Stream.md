---
color: "#5B6CDA"
---

# Stream

按消息 ID 组织的追加式消息流，支持消费者组和消息确认。

## 1. 常见用途

事件流、消息消费、消费者组协作。

## 2. 常用命令

`XADD`、`XREAD`、`XRANGE`、`XGROUP CREATE`、`XREADGROUP`、`XACK`、`XPENDING`、`XTRIM`。

## 3. 底层编码

基数树（rax）与 listpack 的组合。

实现说明以 Redis 7.2 为基线。

[返回数据类型目录](Data Types.md)

