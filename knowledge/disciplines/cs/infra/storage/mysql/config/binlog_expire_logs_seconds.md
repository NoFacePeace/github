---
color: "#5B6CDA"
---

# binlog_expire_logs_seconds

binlog_expire_logs_seconds 指定 [binlog](../logs/Binlog.md) 文件自动过期清理的保留时间，单位为秒。它控制日志保留，不控制每次提交是否刷盘。

## 1. 作用域与默认值

| 属性 | MySQL 8.0 / 8.4 |
| --- | --- |
| 默认值 | 2592000 秒，即 30 天。 |
| 作用域 | GLOBAL。 |
| 动态修改 | 支持，需相应管理权限。 |
| 值为 0 | 不通过该秒数参数设置自动过期时间；旧版本还需检查 expire_logs_days。 |

MySQL 5.7 使用 expire_logs_days，默认 0，表示不按天自动过期。MySQL 8.0 早期版本还需结合该旧参数理解实际策略，不能混用不同版本的配置规则。

## 2. 清理触发与边界

符合过期条件的文件通常在启动、binlog 轮换等时机清理，不保证到期立即删除。降低保留时间也不等于该设置语句会立刻完成所有清理。

MySQL 8.0.29 起支持 binlog_expire_logs_auto_purge；关闭该开关会影响自动过期清理。保留秒数不是磁盘容量上限，写入量大时，即使尚未到期也可能耗尽磁盘。

## 3. 查看与修改

```sql
SELECT @@GLOBAL.binlog_expire_logs_seconds;

-- 通过 SHOW 查询，可用于确认当前版本是否存在该开关。
SHOW GLOBAL VARIABLES LIKE 'binlog_expire_logs_auto_purge';

-- 示例：将保留时间设为 30 天。
SET GLOBAL binlog_expire_logs_seconds = 2592000;

-- MySQL 8.0：修改运行值并持久化。
SET PERSIST binlog_expire_logs_seconds = 2592000;
```

两种修改方式按需求选择，调整前应确认所需日志已归档或仍在保留范围内。日志删除后，仅把保留时间调长不能找回已删除文件。

## 4. 如何选择保留时间

需要同时覆盖副本和 CDC 的消费进度、全量备份周期、时间点恢复目标，以及故障发现和处置所需时间。不能只按正常复制延迟设置很短的保留期。

恢复依赖从可用备份对应位置开始的完整日志链；有归档时，也应确认归档可读取、完整且与备份衔接。自动过期机制不能代替消费进度检查和恢复演练。

应根据日志产生速率评估空间并监控磁盘。确需手动清理时使用 PURGE BINARY LOGS，先确认消费者与恢复需求，不要直接从文件系统删除日志文件。
