# 调度通知频道

用于通知调度器检查等待队列，或在容量变化后重新尝试调度。

## 1. 频道名称

以下使用通用命名空间：

```text
<namespace>:<runnerScope>:runner:capacity-available
```

这是 Pub/Sub 频道，不是持久化数据 key，没有 String、Hash 等存储类型或 TTL。

## 2. 发布与订阅

订阅方通过 `SUBSCRIBE` 监听；入队方通过 `PUBLISH` 发布空字符串通知。

通知不携带任务内容。订阅方收到后，从队列及调度记录读取实际数据。Pub/Sub 不提供离线消息保留，不能替代持久化队列。
