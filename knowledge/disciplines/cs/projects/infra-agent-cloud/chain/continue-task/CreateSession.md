# 创建新 Session

## 1. 处理内容

在已有 Task 下创建并保存新的 [Session](../../storage/mysql/session/Session.md)：

- 新生成本轮 UUID。
- `taskId` 关联已有 Task。
- `parentSessionId` 关联上一轮 Session，没有时为空。
- 保存本轮 prompt、创建者和初始扩展信息。
- 初始状态设为 running。

此处的 running 表示本轮进入执行生命周期，不代表 Runner 已经启动引擎。上一轮即使失败或停止，也不会被复用为本轮记录。

## 2. 下一步

进入[登记回调与开始通知](RegisterCallbacks.md)。
