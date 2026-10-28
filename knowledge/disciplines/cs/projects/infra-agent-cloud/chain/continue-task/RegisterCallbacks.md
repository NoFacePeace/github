# 登记回调与开始通知

## 1. 处理内容

新 Session 保存后，按请求中适用的外部 callback URL 创建本轮回调记录，再发出 Session 开始通知；适用时异步补充任务关注人。

这里是新轮次的回调登记。后续远程执行准备还会登记执行事件回调，不能把外部 callback URL、Redis 回调开关和 SSE 连接理解为同一对象。

## 2. 下一步

进入[更新任务状态](UpdateTaskStatus.md)。
