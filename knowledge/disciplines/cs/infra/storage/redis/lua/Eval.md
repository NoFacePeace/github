---
color: "#5B6CDA"
---

# EVAL

`EVAL` 是 Redis 提供的执行 Lua 脚本的命令。客户端发送脚本与参数，由 Redis 在服务端执行，并将脚本的返回值交给客户端。

## 1. 命令格式

```text
EVAL script numkeys [key ...] [arg ...]
```

| 参数 | 含义 |
| --- | --- |
| `script` | Lua 脚本内容。 |
| `numkeys` | 后续参数中，键名的数量；没有键时传 `0`。 |
| `key ...` | 脚本访问的 Redis 键，通过 `KEYS` 数组读取。 |
| `arg ...` | 其他数据参数，通过 `ARGV` 数组读取。 |

`KEYS` 和 `ARGV` 的下标都从 `1` 开始。不要把这里的 `KEYS` 数组与 Redis 的 `KEYS` 命令混淆。

## 2. 最小示例

```redis
SET stock 10
EVAL "return redis.call('GET', KEYS[1])" 1 stock
```

执行过程：

1. `EVAL` 启动 Lua 脚本。
2. `numkeys` 为 `1`，因此 `stock` 被放入 `KEYS[1]`。
3. `redis.call('GET', KEYS[1])` 在 Redis 中执行 `GET stock`。
4. `return` 将读到的字符串 `10` 返回给客户端。

`EVAL` 是脚本入口，`redis.call()` 是脚本内部执行 Redis 命令的接口。脚本返回值由 Redis 转换为协议响应，例如字符串、整数或数组；不要依赖 Lua 浮点数直接返回时保留小数，可显式返回字符串。

## 3. 传入键与数据

```redis
EVAL "return redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])" 1 greeting hello 60
```

参数对应关系：

```text
KEYS[1] = greeting
ARGV[1] = hello
ARGV[2] = 60
```

脚本将 `greeting` 设置为 `hello`，并设置 60 秒过期时间。参数传入 Lua 时是字符串，需要数值计算时可用 `tonumber()` 转换并校验范围。

脚本访问的所有键都应显式放在键参数中，不应通过 `ARGV` 传递键名，或根据数据内容生成未声明的键。数据通过参数传入，避免将用户输入直接拼进脚本。

## 4. 多命令与条件判断

例如释放锁时，需要先确认 token 属于自己，再删除锁。脚本内容如下：

```lua
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
    return 0
end

return redis.call('DEL', KEYS[1])
```

可以直接执行：

```redis
EVAL "if redis.call('GET', KEYS[1]) ~= ARGV[1] then return 0 end; return redis.call('DEL', KEYS[1])" 1 task-lock my-token
```

返回 `1` 表示成功删除，`0` 表示键不存在或 token 不匹配。比较 token 与删除之间，其他客户端不能插入命令，避免旧持有者删除已经属于新持有者的锁。

## 5. 客户端调用

使用 ioredis 时，参数顺序与命令格式一致：

```ts
const script = `
  if redis.call('GET', KEYS[1]) ~= ARGV[1] then
    return 0
  end
  return redis.call('DEL', KEYS[1])
`

const deleted = await redis.eval(script, 1, 'task-lock', token)
```

不同客户端或项目封装的函数签名可能不同。例如某些封装使用 `eval(script, keysArray, argsArray)`，内部再转换为 Redis 命令参数，不能直接套用到所有客户端。

## 6. 执行边界

- **执行隔离**：脚本执行期间，其他客户端的普通命令不能插入；适合将读取、判断和修改放在一起完成。
- **没有自动回滚**：脚本后续报错，不会撤销之前成功的写入，应尽量先校验再修改。
- **控制耗时**：脚本会阻塞其他客户端的普通命令，避免长循环、全量扫描或大量计算。
- **集群约束**：Redis Cluster 中，脚本访问的多个键通常必须在同一个哈希槽，可使用相同 hash tag，如 `task:{123}:record` 与 `task:{123}:claim`。
- **重试边界**：客户端超时不代表脚本未执行；重试修改类脚本时，需要考虑幂等性。
- **脚本复用**：脚本保持固定，将变化的数据放入 `KEYS` 和 `ARGV`。需要按摘要调用时，可结合 `SCRIPT LOAD` 与 `EVALSHA`，并处理脚本缓存丢失后的 `NOSCRIPT` 错误。
