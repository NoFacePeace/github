# SessionRunnerMapping

保存某轮执行对应的 Runner 和租约，作为派发归属记录。这里采用通用实体名称，不记录内部项目实际表名。

## 1. 主要字段

| 字段 | 含义 |
| --- | --- |
| id | 自增主键。 |
| task_uuid | 所属任务 UUID。 |
| session_uuid | 本轮执行 UUID，具有唯一约束。 |
| runner_scope | Runner 资源池范围。 |
| runner_id | 分配的 Runner。 |
| lease_id | 本次执行租约标识。 |
| created_at / updated_at | 创建与更新时间。 |

## 2. 写入时机

准备结果已经保存为 `dispatching` 后、向 Job Stream 投递之前插入。同一 Session 的重试必须与已有 Task、Runner 及租约信息一致，否则判定为冲突。

## 3. 与其他记录的区别

本表记录执行归属，不保存完整执行输入，也不是任务进度表。Redis 执行租约提供带有效期的执行协调，本表提供持久化关联；记录存在不代表 Agent 已启动。
