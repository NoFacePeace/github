# 加入等待队列

将本轮 Session 加入 Redis 的[待调度集合](../../../../../storage/redis/DispatchPending.md)，供调度器查找等待执行的请求。

## 1. 队列元素

- 成员：Session UUID。
- 分数：排队时间的毫秒时间戳。

使用 Sorted Set 按排队时间组织请求。集合中不保存完整执行输入，调度器根据 Session UUID 读取对应调度记录。

## 2. 写入方式

通过 `ZADD` 添加成员，与保存调度记录在同一 Lua 脚本中执行。

进入等待队列仅表示等待调度，尚未取得 Runner 执行权。
