---
color: "#5B6CDA"
---

# MySQL 配置

MySQL 配置控制连接、内存、日志、事务持久性及运行行为。参数的默认值、作用域、动态修改能力和持久化支持随版本变化，应以实例实际值与对应版本规则为准。

## 1. 配置来源

| 来源 | 说明 |
| --- | --- |
| 内置默认值 | 未被其他来源覆盖时使用的值。 |
| 选项文件 | 如 my.cnf、my.ini，服务端选项通常写在 [mysqld] 分组中。 |
| 启动参数 | 由启动命令或服务管理配置传入。 |
| 持久化变量文件 | MySQL 8.0 的 SET PERSIST 等操作维护 mysqld-auto.cnf。 |
| 运行时设置 | 通过 SET GLOBAL 或 SET SESSION 修改允许动态调整的变量。 |

配置文件路径、加载顺序和覆盖关系取决于平台、版本及启动方式。不能只修改一份 my.cnf 就认定实例一定读取了它；mysqld-auto.cnf 应通过对应 SQL 管理，不应手动编辑。

## 2. 查看全局与会话配置

```sql
SHOW GLOBAL VARIABLES LIKE 'innodb_flush_log_at_trx_commit';
SHOW GLOBAL VARIABLES LIKE 'sync_binlog';
SHOW SESSION VARIABLES LIKE 'transaction_isolation';

SELECT @@GLOBAL.max_connections,
       @@SESSION.transaction_isolation;
```

GLOBAL 影响实例级行为，或作为新会话的初始默认值；SESSION 只影响当前连接。并非所有变量都同时支持两种作用域。

对具有会话值的变量，修改全局默认值通常不会同步改变已经建立的连接。连接池中的旧连接可能继续沿用原会话值。

## 3. 动态修改与持久化

以下语句用于说明机制，并非推荐直接应用的调优值：

```sql
-- 仅修改当前连接支持的会话变量。
SET SESSION long_query_time = 2;

-- 修改全局运行值，单独使用不会保存到重启后的配置。
SET GLOBAL long_query_time = 2;

-- MySQL 8.0：修改全局运行值并保存持久化设置。
SET PERSIST long_query_time = 2;
```

SET PERSIST_ONLY 保存持久化设置而不改变当前运行值，可用于支持此机制的只读变量等场景；具体变量限制与权限需结合版本确认。

静态参数通常需要通过启动配置并重启生效。修改普通选项文件也不会自动改变当前运行值；动态参数需要显式在线设置，或在后续启动时读取。

## 4. 常见参数

| 类别 | 参数示例 | 关注点 |
| --- | --- | --- |
| 连接 | max_connections | 并发连接上限，不等于数据库可高效处理的并发量。 |
| 内存 | innodb_buffer_pool_size | 数据与索引页缓存，需要给连接、排序、系统及其他进程预留空间。 |
| 事务 | transaction_isolation | 隔离级别影响可见性与锁行为，需区分全局默认和会话实际值。 |
| redo 持久性 | [innodb_flush_log_at_trx_commit](innodb_flush_log_at_trx_commit.md) | 控制提交相关 redo 写入和刷盘策略。 |
| binlog 持久性 | [sync_binlog](sync_binlog.md) | 控制 binlog 同步频率，影响性能与故障下的持久性。 |
| binlog 保留 | [binlog_expire_logs_seconds](binlog_expire_logs_seconds.md) | 控制过期时间，需覆盖复制和恢复所需日志链。 |
| 慢查询采集 | slow_query_log、long_query_time | 控制慢日志启用和耗时阈值，还需结合其他记录条件。 |
| 数据锁等待 | innodb_lock_wait_timeout | 控制 InnoDB 行锁等待超时，不是通用 SQL 执行超时。 |

## 5. 变更验证

修改前记录版本、当前值、来源、期望作用域和恢复方式，明确是否需要重启或重建连接。持久性参数还应结合故障容忍目标评估，不能只比较正常运行时的吞吐。

修改后分别验证运行值、代表性已有会话、新会话及必要时的重启结果，并比较延迟、资源和错误变化。运行值改变不等于持久化设置已保存，持久化设置存在也不等于当前值已经生效。
