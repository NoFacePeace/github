# 派发

准备派发成功、调度记录已保存为 `dispatching` 后，使用记录中已提交的 Job 向目标 Runner 投递作业。

## 1. 处理顺序

从[记录执行归属](RecordSessionRunner.md)开始，沿后续链接处理：

```text
记录 Session 与 Runner 的关联
  → 去重并写入 Job Stream
  → 保存派发结果
  → 同步业务状态并确认结果
```

## 2. 阶段边界

本阶段写 MySQL 关联记录和 Redis 消息、状态，不直接上传对象存储文件。

`dispatching` 表示已有可下发的 Job，`dispatched` 表示投递步骤成功，不代表 Runner 已接收、Agent 已启动或执行完成。
