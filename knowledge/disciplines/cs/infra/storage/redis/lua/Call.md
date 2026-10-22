---
color: "#5B6CDA"
---

# redis.call()

`redis.call()` 是 Redis 内置 Lua 环境提供的函数，用于在脚本中执行 Redis 命令，并将命令结果返回给 Lua。`EVAL` 负责启动脚本，`redis.call()` 负责在脚本内部操作数据。

## 1. 调用格式

```lua
redis.call(command, argument1, argument2, ...)
```

第一个参数是命令名称，后面依次传入命令需要的参数。命令名称不区分大小写，通常使用大写以便阅读。

```lua
local value = redis.call('GET', KEYS[1])
redis.call('SET', KEYS[1], ARGV[1])
redis.call('EXPIRE', KEYS[1], 60)
```

`KEYS` 和 `ARGV` 都从 `1` 开始索引，分别接收客户端传入的键名和其他参数。脚本访问的键应由客户端显式声明在 `KEYS` 中。

不要把整条命令拼成一个参数：

```lua
-- 正确：命令名和各参数分别传入。
redis.call('SET', KEYS[1], ARGV[1])

-- 错误：Redis 会把整个字符串当作命令名称。
redis.call('SET greeting hello')
```

## 2. 通过 EVAL 执行

```redis
EVAL "return redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])" 1 greeting hello 60
```

对应的脚本是：

```lua
return redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])
```

它执行 `SET greeting hello EX 60`。`redis.call()` 把结果返回给 Lua，外层的 `return` 再将结果返回给客户端。

EVAL 的参数格式与客户端调用方式见 [EVAL](Eval.md)。

## 3. 常用命令

| Lua 调用 | 作用 |
| --- | --- |
| `redis.call('GET', KEYS[1])` | 读取字符串。 |
| `redis.call('SET', KEYS[1], ARGV[1])` | 写入字符串。 |
| `redis.call('DEL', KEYS[1])` | 删除键，返回删除数量。 |
| `redis.call('EXISTS', KEYS[1])` | 检查键是否存在，单键调用返回 `0` 或 `1`。 |
| `redis.call('EXPIRE', KEYS[1], 30)` | 设置 30 秒过期时间。 |
| `redis.call('ZADD', KEYS[1], ARGV[1], ARGV[2])` | 添加或更新有序集合成员及分数。 |
| `redis.call('ZRANGE', KEYS[1], 0, -1)` | 按分数从小到大读取全部成员。 |
| `redis.call('ZREM', KEYS[1], ARGV[1])` | 移除指定成员。 |
| `redis.call('ZCARD', KEYS[1])` | 获取有序集合成员数量。 |

## 4. 返回值与条件判断

在默认的 RESP2 脚本响应转换下，常见命令结果会映射为以下 Lua 值；调用 `redis.setresp(3)` 后，部分转换规则不同。

| Redis 命令结果 | Lua 中的值 |
| --- | --- |
| 字符串，例如 `GET` 读取到的内容 | 字符串。 |
| 整数，例如 `EXISTS` 返回的数量 | 数字。 |
| 数组，例如 `ZRANGE` 返回的成员列表 | Lua table，通常从 `1` 开始索引。 |
| 空响应，例如不存在的键的 `GET` 结果 | `false`。 |
| 状态响应，例如 `SET` 的 `OK` | 含 `ok` 字段的 table，例如 `{ ok = 'OK' }。 |

Lua 中只有 `false` 和 `nil` 被视为假，数字 `0` 也被视为真。因此，不能直接用 `if redis.call('EXISTS', ...) then` 判断存在性，应显式比较：

```lua
if redis.call('EXISTS', KEYS[1]) == 1 then
    return redis.call('GET', KEYS[1])
end

return false
```

读取字符串后需要计算时，应先进行数值转换与校验；`GET` 得到的字符串 `'0'` 与数字 `0` 不相等。

## 5. 错误处理与 pcall 的区别

`redis.call()` 执行命令报错时，若没有在 Lua 层捕获，脚本立即中止并向客户端返回错误。之前已经成功的写入不会自动回滚。

```lua
redis.call('SET', KEYS[1], 'hello')
redis.call('LPUSH', KEYS[1], 'item') -- WRONGTYPE，脚本在此中止。
redis.call('SET', KEYS[2], 'done')   -- 不会执行。
```

执行后，`KEYS[1]` 仍保存字符串 `hello`。

如果希望检查错误并继续处理，可以使用 `redis.pcall()`：

```lua
local result = redis.pcall('LPUSH', KEYS[1], ARGV[1])
if type(result) == 'table' and result.err then
    return { 0, result.err }
end

return { 1, result }
```

`redis.pcall()` 在命令失败时返回含 `err` 字段的 table；它只改变错误的处理方式，不提供自动回滚。

## 6. 执行边界

- 原子执行的范围是整段 Lua 脚本；多个 `redis.call()` 之间不会插入其他客户端的普通命令。
- `redis.call()` 不能绕过脚本执行限制，例如不能在脚本内使用 `MULTI` 开启事务，也不能依赖阻塞式取队列命令在脚本中等待新消息。
- Redis Cluster 中，脚本访问的多个键通常必须位于同一哈希槽。
- 尽量先检查参数与业务条件，再执行写操作；控制脚本耗时，避免大量扫描和循环。
