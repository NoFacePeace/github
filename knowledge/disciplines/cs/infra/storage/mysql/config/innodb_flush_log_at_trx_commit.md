---
color: "#5B6CDA"
---

# innodb_flush_log_at_trx_commit

innodb_flush_log_at_trx_commit 控制事务提交时 InnoDB [redo log](../innodb/logs/RedoLog.md) 的写入与刷盘要求。它影响日志持久性，不表示每次提交必须将所有修改的数据页写回数据文件。

## 1. 作用域与默认值

| 属性 | MySQL 8.0 / 8.4 |
| --- | --- |
| 默认值 | 1。 |
| 作用域 | GLOBAL。 |
| 动态修改 | 支持，需相应管理权限。 |
| 常用取值 | 0、1、2。 |

实例配置可能覆盖默认值，应查询实际设置。

## 2. 不同取值

| 值 | 提交时的要求 | 故障风险 |
| --- | --- | --- |
| 1 | 写出并持久化相关 redo；实际 I/O 可以通过组提交合并。 | 在存储正确兑现同步请求等前提下，提供提交持久性保障。 |
| 2 | 写出相关 redo 到操作系统缓存，不要求每次提交同步刷盘。 | 操作系统崩溃或掉电时可能丢失尚未刷盘的已提交日志。 |
| 0 | 提交本身不要求立即写出和刷盘，通常由后台周期性处理。 | mysqld 进程崩溃就可能丢失仍在日志缓冲区中的已提交日志。 |

周期性写入和刷盘的间隔还受 innodb_flush_log_at_timeout、线程调度和其他刷新触发条件影响。不能把潜在丢失量理解为严格固定的一秒，也不能认为非提交时刻完全不会刷日志。

## 3. 查看与修改

```sql
SELECT @@GLOBAL.innodb_flush_log_at_trx_commit;

-- 修改运行值，不单独保存重启后的设置。
SET GLOBAL innodb_flush_log_at_trx_commit = 1;

-- MySQL 8.0：修改运行值并持久化。
SET PERSIST innodb_flush_log_at_trx_commit = 1;
```

以上为配置机制示例，不需要依次执行两种设置。修改会影响全局事务持久性策略，应根据允许的数据丢失范围与性能证据决定。

## 4. 与 binlog 的配合

开启 binlog 时，还需考虑 [binlog](../logs/Binlog.md) 的同步策略。常用强持久性组合为 innodb_flush_log_at_trx_commit = 1、sync_binlog = 1。

仅保证 redo 刷盘，不代表 binlog 也已持久化；两阶段提交的协调能力不能替代日志同步和可靠存储。
