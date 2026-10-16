---
color: "#5B6CDA"
---

# Enable at Runtime

运行时开启 AOF，是指 Redis 进程已经启动，原本未开启 AOF 持久化，此时通过命令开启，而不重启实例。

## 1. 触发方式

Redis 已运行且原本关闭 AOF 时，可以动态开启：

```redis
CONFIG SET appendonly yes
```

此时内存可能已经有大量数据。只记录开启后的增量操作无法恢复原有数据，所以 Redis 使用后台重写机制生成包含现有数据的基础内容，同时记录新的写操作。

## 2. 与自动重写的区别

开启 AOF 与开启自动重写是两个不同的设置：

| 配置项 | 作用 |
| --- | --- |
| `appendonly` | 控制是否开启 AOF 持久化。 |
| `auto-aof-rewrite-percentage` | 控制按文件增长比例触发的自动重写，设为 `0` 可以关闭这类自动重写。 |

即使关闭了按增长阈值触发的自动重写，运行时开启 AOF 仍需要生成包含现有数据的基础内容。

## 3. 确认开启结果

开启后，应通过 `INFO persistence` 确认后台重写成功，不能仅凭 `CONFIG SET` 返回成功就认为基础文件已经生成完毕。

```redis
CONFIG GET appendonly
INFO persistence
```

`appendonly` 应为 `yes`；结合 `aof_rewrite_in_progress`、`aof_rewrite_scheduled` 和 `aof_last_bgrewrite_status` 判断本次重写是否完成并成功，同时检查 `aof_last_write_status` 是否正常。

## 4. 重启后是否保留

`CONFIG SET appendonly yes` 只修改当前进程的配置，不会自动修改启动配置。如果启动配置仍为 `appendonly no`，下次重启仍会关闭 AOF。

要让重启后继续开启，需要将启动时使用的配置文件设置为：

```conf
appendonly yes
```

也可以在运行时开启并确认成功后，执行以下命令，将当前配置写回启动时使用的配置文件：

```redis
CONFIG REWRITE
```

未使用配置文件启动或文件不可写时，`CONFIG REWRITE` 会失败。容器或托管环境还需要同步更新部署配置来源，避免实例重建时恢复旧配置；启动参数也应与预期配置一致。

在已有数据的实例中切换到 AOF 时，应先完成运行时开启并确认基础文件生成成功，再保存配置。不要只把配置文件改成 `appendonly yes` 后直接重启，并假设已有数据已经写入 AOF。
