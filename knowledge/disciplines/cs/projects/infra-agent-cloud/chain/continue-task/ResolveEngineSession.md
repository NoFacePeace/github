# 查找引擎会话标识

## 1. 处理内容

优先使用上一轮 Session 的 `executorSessionId`。若为空，再查询该 Task 中按创建时间倒序、该字段非空的最近 Session。

这是读取 MySQL 中保存的引擎会话标识，不是从日志或对象存储文件名推导。全部没有时使用空值，后续启动新引擎会话。

适用执行路径会将该标识作为 `resumeSessionId` 传递给 Executor，用来恢复引擎对话；它不同于平台新建的 Session UUID，也不能恢复旧进程。

## 2. 下一步

进入[创建新 Session](CreateSession.md)。
