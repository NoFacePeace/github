# 检查执行租约

## 1. 处理内容

未取消排队时，检查本轮是否已有[执行租约](../../storage/redis/ExecutionLease.md)。有租约才进入远程停止命令发布。

暂时没有租约时，不在这里认定 Session 已停止；停止标记仍保存在 MySQL，后续提交、派发结果或恢复处理会检查并再次尝试停止。

## 2. 后续流程

有租约则[定位并校验 Runner](ResolveRunnerOwner.md)；没有租约则返回接口侧[更新 Task 状态](UpdateTaskStatus.md)，等待后续处理停止意图。
