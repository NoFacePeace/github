# 读取等待队列

调度器开始处理队列时，先读取 Redis 中的[待调度集合](../../../../storage/redis/DispatchPending.md)，取得等待处理的 Session UUID。

## 1. 读取方式

```text
ZRANGE <待调度集合 key> 0 -1
```

按分数从小到大读取所有成员。集合以排队时间为分数，因此通常先检查较早入队的请求。

集合只返回 Session UUID，不包含完整执行输入，也不表示这些请求已经获得 Runner。

## 2. 存储操作

这一步只读取 Redis，不写 MySQL，也不移除集合成员。多个 Server 可以读到同一个 Session UUID，读取本身不赋予调度处理权。

## 3. 后续步骤

遍历读取到的 Session UUID，逐条尝试[获取调度处理权](AcquireDispatchClaim.md)。成功后才读取对应调度记录、检查状态与排队超时，并尝试分配 Runner；未取得处理权时跳过该项。

集合按排队时间排序不保证实际执行顺序，节点能力、容量及并发争用仍会影响分配结果。
