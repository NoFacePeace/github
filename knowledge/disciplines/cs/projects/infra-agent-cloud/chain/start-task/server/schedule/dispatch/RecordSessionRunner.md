# 记录执行归属

使用已保存 Job 中的任务、执行轮次、Runner 与租约信息，在 MySQL 的 [SessionRunnerMapping](../../../../../storage/mysql/session-runner-mapping/SessionRunnerMapping.md) 表插入执行归属记录。

## 1. 幂等检查

Session UUID 具有唯一约束。重复插入时，读取已有记录，核对 Task、Runner 范围、Runner ID 与租约 ID；全部一致可继续，否则返回归属冲突。

仅在 Job 已经提交到调度记录后才写关联，避免尚未固定的准备结果占用唯一关联。

## 2. 后续步骤

进入[投递 Job](PublishJob.md)。关联记录保存成功不代表 Stream 消息已经写入。
