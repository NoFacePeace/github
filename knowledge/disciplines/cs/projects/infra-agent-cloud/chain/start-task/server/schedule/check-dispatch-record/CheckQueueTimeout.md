# 检查排队超时

确认记录仍处于 `queued` 后，比较当前时间与调度记录中的 [deadline](../../../../../storage/redis/dispatch-record/Deadline.md)。

## 1. 已超时

```text
当前时间 >= deadline
```

将调度状态改为 `failed`，记录排队超时原因，并在校验 claim 令牌后保存到 Redis。该分支不继续分配 Runner。

后续需成功同步业务状态并确认调度结果，才移除等待成员并为调度记录设置清理 TTL。

## 2. 未超时

进入[准备派发](../prepare-dispatch/PrepareDispatch.md)，解密执行输入、查找可用 Runner 并尝试建立执行归属。容量不足等可恢复情况可能使请求继续等待。

## 3. 与 Redis TTL 的区别

`deadline` 是业务超时判断依据，到达该时间不会自动删除 Redis key。排队超时由调度逻辑处理，与调度结果确认后设置的 24 小时 TTL 分开管理。
