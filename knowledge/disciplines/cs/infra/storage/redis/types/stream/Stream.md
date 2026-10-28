---
color: "#5B6CDA"
---

# Stream

按消息 ID 组织的追加式消息流，支持消费者组和消息确认。

## 1. 数据模型

消息记录、消费者组、投递进度与待确认列表的关系见 [Model](model/Model.md)。

## 2. 常见用途

事件流、消息消费、消费者组协作。

## 3. 常用命令

`XADD`、`XREAD`、`XRANGE`、`XGROUP CREATE`、`XREADGROUP`、`XACK`、`XPENDING`、`XTRIM`。

命令用法与示例见 [Commands](commands/Commands.md)。

## 4. 底层编码

基数树（rax）与 listpack 的组合。
