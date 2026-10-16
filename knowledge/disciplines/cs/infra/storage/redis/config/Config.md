---
color: "#5B6CDA"
---

# Config

Redis 可以通过配置文件、启动参数和运行时命令设置行为。实际生效的配置取决于启动方式和后续修改。

## 1. 配置文件

Redis 配置文件通常命名为 `redis.conf`，每行指定一个配置项及其参数，使用 `#` 添加注释。

以下示例开启 AOF，并设置每秒同步一次：

```conf
appendonly yes
appendfsync everysec
```

启动时显式指定配置文件：

```sh
redis-server /path/to/redis.conf
```

修改文件后，正在运行的 Redis 不会自动重新加载。需要重启实例，或对支持动态修改的配置项使用 `CONFIG SET`。

## 2. 启动参数

启动参数可以覆盖配置文件中的同名配置项，例如：

```sh
redis-server /path/to/redis.conf --appendonly yes --appendfsync everysec
```

## 3. 运行时配置

使用 `CONFIG GET` 查询当前实例实际生效的值：

```redis
CONFIG GET appendonly
CONFIG GET appendfsync
```

使用 `CONFIG SET` 修改支持动态调整的配置项，例如调整 AOF 同步策略：

```redis
CONFIG SET appendfsync everysec
```

运行时修改不会自动保存到配置文件。需要更新启动时使用的配置文件，或执行：

```redis
CONFIG REWRITE
```

`CONFIG REWRITE` 将当前配置写回启动时指定的配置文件；未使用配置文件启动或文件不可写时会失败。容器或托管环境还需要按部署方式更新配置来源，避免实例重建后恢复旧配置。

## 4. AOF 配置

| 配置项 | Redis 默认值 | 作用 |
| --- | --- | --- |
| [appendonly](appendonly.md) | `no` | 是否开启 AOF 持久化。 |
| [appendfsync](appendfsync.md) | `everysec` | AOF 同步策略，支持 `always`、`everysec` 和 `no`；开启 AOF 后生效。 |

部署环境可能覆盖上述默认值，应使用 `CONFIG GET` 确认当前实例的实际配置。

## 5. 内存配置

| 配置项 | 常见 64 位 Redis 默认值 | 作用 |
| --- | --- | --- |
| [maxmemory](maxmemory.md) | `0` | Redis 的内存上限，`0` 表示不设置此上限。 |
| [maxmemory-policy](maxmemory-policy.md) | `noeviction` | 达到上限时淘汰哪些键，或拒绝需要增加内存的写入。 |
