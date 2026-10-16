---
color: "#5B6CDA"
---

# MySQL 日志

MySQL 日志分布在 Server 层和存储引擎层，分别服务于运行诊断、查询分析、复制、恢复与事务管理。不同日志的记录内容、生命周期和持久化方式不同。

## 1. Server 层日志

| 日志 | 主要内容与用途 |
| --- | --- |
| 错误日志（Error Log） | 记录启动、关闭、错误和诊断消息等，用于排查运行问题。 |
| 通用查询日志（General Query Log） | 记录客户端连接及收到的语句，用于观察访问活动；不等同于事务成功提交记录。 |
| 慢查询日志（Slow Query Log） | 按耗时阈值及相关配置记录查询，帮助分析性能。 |
| 二进制日志（Binary Log，binlog） | 以事件记录数据变更等信息，用于复制和结合备份进行按时间点恢复。 |
| 中继日志（Relay Log） | 副本接收源端 binlog 事件后暂存的日志，供复制应用线程处理。 |

通用查询日志和慢查询日志的输出位置、过滤规则及开销取决于配置。binlog 不是普通的 SQL 文本访问日志，记录形式受日志格式等因素影响。

## 2. InnoDB 日志

| 日志 | 主要用途 |
| --- | --- |
| redo log | 记录页修改相关信息，通过预写日志机制支持崩溃恢复。 |
| undo log | 保存撤销修改所需的信息，用于事务回滚，并支持重建 MVCC 历史版本。 |

redo、undo 是存储引擎机制，不能推广为所有 MySQL 存储引擎都具有相同实现。

undo 记录采用专门的格式，通常保存在 undo 表空间中，不是使用 DYNAMIC 行格式保存完整旧行快照。redo 也不是面向人工阅读的查询文本日志。

## 3. binlog、redo 与 undo 的分工

| 需求 | 相关机制 |
| --- | --- |
| 将变更复制到副本 | binlog，以及副本上的 relay log。 |
| 从备份恢复到指定时间点 | 备份与后续 binlog。 |
| 崩溃后恢复 InnoDB 数据状态 | redo、undo 及事务恢复协调机制。 |
| 撤销未提交修改 | undo。 |
| 一致性读重建历史版本 | undo、记录事务信息与 Read View 等。 |

启用 binlog 时，MySQL 通过内部两阶段提交协调 InnoDB 与 binlog 的事务提交状态。持久性还受日志刷盘参数、存储设备等条件影响，不能只根据“开启日志”就推断全部保障。

## 4. 日志与数据页刷盘

修改通常先发生在 Buffer Pool 页中，相关 redo 按预写日志规则先于对应数据页持久化。事务提交不要求把其修改的所有数据页立即写回数据文件。

日志刷新、日志轮换和数据页写回是不同过程。FLUSH LOGS 对不同日志执行相应处理，不等于将所有脏页写盘，也不等于删除历史日志。

日志不是统一的“记录所有操作并永久保留”的机制。诊断日志可能未启用，binlog 有过期清理策略，redo 空间会复用，undo 历史则会在不再需要时清理。

## 5. 查看配置

```sql
SHOW VARIABLES LIKE 'log_error';
SHOW VARIABLES LIKE 'general_log';
SHOW VARIABLES LIKE 'slow_query_log';
SHOW VARIABLES LIKE 'long_query_time';
SHOW VARIABLES LIKE 'log_bin';
SHOW VARIABLES LIKE 'binlog_format';
SHOW VARIABLES LIKE 'relay_log';
SHOW VARIABLES LIKE 'innodb_flush_log_at_trx_commit';
SHOW VARIABLES LIKE 'sync_binlog';
```

变量、日志默认行为与输出方式可能随版本变化。启用高流量访问日志或调整刷盘、保留策略前，应评估性能、空间及恢复需求。
