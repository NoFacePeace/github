# 派发去重标记

记录某个 Session 已投递的消息标识，防止重复写入目标 Runner 的 Job Stream。

## 1. Key 与类型

以下采用通用命名空间：

```text
{<namespace>:<runnerScope>:runner:jobs:<runnerId>}:published:<replyTo>:<sessionUuid>
```

花括号包围完整的目标 Stream key 字符串。`replyTo` 表示事件回复的 Server 范围。

类型为 String，值为 Job 的 `messageId`。它与 Job Stream、调度记录、调度 claim 都是独立 key。

## 2. 去重规则

- 标记不存在：检查租约截止时间，通过后追加 Job 并保存标记。
- 标记存在且消息标识相同：确认此前已发布，不重复追加。
- 标记存在但消息标识不同：判定为投递冲突。

这些判断与 Stream 写入由同一 Lua 脚本完成。该标记用于发布去重，不代表作业执行具有端到端的恰好一次保证。

## 3. 有效期

首次保存时，通过 `SET PX` 设置有效期：

```text
TTL = max(24 小时, Job 租约截止时间 - Redis 当前时间)
```

因此初始 TTL 至少为 24 小时。命中已有相同标记时直接返回，不在该分支刷新 TTL。
