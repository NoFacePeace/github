# 接收停止请求

## 1. 处理内容

页面点击 Stop 并确认后，调用 `POST /tasks/:taskUuid/stop`。Server 校验任务操作权限，再进入任务服务。

本链路以远程沙箱 Task 为例；指定单个 Session 的停止接口复用后续能力，但权限校验和目标范围有所不同。停止不会创建新 Session，也不删除历史对话。

## 2. 后续流程

[查询运行中的 Session](QueryRunningSessions.md)。
