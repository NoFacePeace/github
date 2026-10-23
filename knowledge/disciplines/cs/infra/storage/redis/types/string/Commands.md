---
color: "#5B6CDA"
---

# Commands

String 命令用于读写字符串、批量访问、数字增减及字节范围操作。String 是二进制安全的字节序列，Redis 不会自动理解其中的 JSON 或其他业务格式。

## 1. 命令概览

| 命令 | 作用 |
| --- | --- |
| `SET` | 设置值，可指定条件和过期策略。 |
| `GET` | 读取一个字符串值。 |
| `MSET`、`MGET` | 批量写入或读取。 |
| `SETNX`、`MSETNX` | 仅在目标 key 不存在时写入。 |
| `GETDEL` | 读取并删除字符串值。 |
| `GETEX` | 读取值，同时修改过期策略。 |
| `INCR`、`DECR` | 整数加一或减一。 |
| `INCRBY`、`DECRBY` | 按指定整数增减。 |
| `INCRBYFLOAT` | 按浮点增量修改数值。 |
| `APPEND` | 在值尾部追加字节。 |
| `STRLEN` | 查询值的字节长度。 |
| `GETRANGE`、`SETRANGE` | 读取或覆盖指定字节范围。 |

## 2. 单值读写

```redis
SET greeting hello
GET greeting
SET greeting world EX 60
SET task-lock unique-token NX PX 30000
```

普通 `SET` 成功时返回 `OK`，会覆盖原值，且默认清除旧 TTL；可以显式指定新的过期策略或使用 `KEEPTTL` 保留已有 TTL。`SET` 能覆盖其他类型的 key，不会因为原值是集合就拒绝写入。

`GET` 返回字符串，key 不存在时返回空响应，key 类型不是 String 时返回 `WRONGTYPE`。没有使用 `GET` 选项的条件 `SET`，条件不满足时返回空响应。

```redis
SETNX task-lock unique-token
```

`SETNX` 成功返回 `1`，key 已存在返回 `0`，本身不设置 TTL。需要临时锁时应使用单条 `SET ... NX PX ...`，避免先设置值再设置有效期时中途失败。

## 3. 批量读写

```redis
MSET user:1 alice user:2 bob
MGET user:1 user:2 user:missing
MSETNX resource:a token-a resource:b token-b
```

`MSET` 在一条命令中写入多个 key，返回 `OK`，覆盖时会清除相应旧 TTL，不支持为每个 key 指定 TTL。

`MGET` 按传入 key 的顺序返回结果；key 不存在或其值不是 String 时，对应位置返回空值，不会像单键 `GET` 那样因非 String key 返回类型错误。批量读取可以减少网络往返，但仍应控制 key 数量和响应大小。

`MSETNX` 仅在所有目标 key 都不存在时写入全部值，成功返回 `1`；任一 key 已存在时不写入任何目标，返回 `0`。它也不设置 TTL。

Redis Cluster 中，多 key 命令通常要求目标位于同一哈希槽，不能直接用它们读取或修改任意分片上的 key。

## 4. 读取并修改生命周期

```redis
GETDEL one-time-token
GETEX cache:item EX 60
GETEX cache:item PERSIST
```

`GETDEL` 原子读取并删除，key 不存在时返回空响应，类型错误时报错。它适合取走一次性值，但不负责处理取走后进程崩溃的恢复问题。

`GETEX` 读取字符串的同时设置或清除 TTL，`PERSIST` 表示移除过期时间。普通 `GET` 不刷新 TTL。`GETDEL` 和 `GETEX` 从 Redis 6.2 起提供。

## 5. 数字增减

以下示例假设 `counter` 初始不存在：

```redis
INCR counter
INCRBY counter 9
DECR counter
DECRBY counter 2
```

返回值依次为 `1`、`10`、`9`、`7`。key 不存在时按初始值 `0` 计算并创建 String；已有值必须是 Redis 能解析的整数，结果必须在有符号 64 位整数范围内，否则报错。

```redis
SET temperature 20
INCRBYFLOAT temperature 0.5
```

`INCRBYFLOAT` 用于浮点增量，不能把浮点运算当作严格十进制金额计算。需要精确金额时，可以根据范围约束使用整数最小单位。

这些增减命令在已有 key 上保留 TTL。单条增减是原子的，但“读取、在客户端计算、再 SET”不是同一原子操作，可能丢失并发更新。

## 6. 字节追加与范围

以下示例假设 `text` 初始不存在：

```redis
APPEND text hello
APPEND text " world"
STRLEN text
GETRANGE text 0 4
SETRANGE text 6 Redis
GET text
```

`APPEND` 返回追加后的字节长度；`STRLEN` 返回 `11`；`GETRANGE` 返回 `hello`；`SETRANGE` 后的值为 `hello Redis`。

范围下标以字节为单位，从 `0` 开始。`GETRANGE` 包含起止位置，支持负下标；`SETRANGE` 使用非负偏移，超出原长度的间隙会用零字节填充。UTF-8 字符可能占多个字节，按字节截取可能切断字符。

不存在的 key 上，`STRLEN` 返回 `0`，`GETRANGE` 返回空字符串。`APPEND` 和 `SETRANGE` 在已有 key 上保留 TTL。

## 7. 通用键命令与边界

`DEL`、`EXPIRE`、`PEXPIRE`、`TTL`、`PTTL`、`EXISTS` 是通用 key 命令，不仅适用于 String。TTL 管理整个 key，不会自动停止使用该值的客户端代码。

字符串存放 JSON 时，普通 String 命令只能整体读写或按字节操作，不能直接按 JSON 字段修改。请求超时也不能证明写入未执行，特别是计数和追加操作，重试前应考虑重复效果。
