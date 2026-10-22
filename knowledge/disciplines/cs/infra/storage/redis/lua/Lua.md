---
color: "#5B6CDA"
---

# Lua

Lua 是一种轻量的脚本语言。Redis 内置 Lua 执行环境，可以在服务端读取数据、判断条件并执行修改，减少客户端与 Redis 之间的往返。

## 1. 调用方式

使用 `EVAL` 执行脚本，依次提供脚本内容、键的数量、键名及其他参数：

详细命令格式、参数与客户端示例见 [EVAL](Eval.md)。

```redis
EVAL "return redis.call('GET', KEYS[1])" 1 stock
```

| 接口 | 作用 |
| --- | --- |
| `KEYS` | 客户端显式传入的键名数组，从 `1` 开始索引。 |
| `ARGV` | 客户端传入的非键名参数数组，从 `1` 开始索引。 |
| [redis.call()](Call.md) | 调用 Redis 命令；命令报错会中止脚本，除非在 Lua 层捕获。 |
| `redis.pcall()` | 调用 Redis 命令；命令报错时返回包含 `err` 字段的错误对象，供脚本继续处理。 |

脚本访问的键应通过 `KEYS` 显式传入，不应根据数据内容动态生成未声明的键名。Redis Cluster 中，通常要求一次脚本访问的键位于同一个哈希槽。

## 2. 条件判断与修改

以下示例先检查扣减数量及库存，再修改库存：

```lua
local amount = tonumber(ARGV[1])
if not amount or amount <= 0 or amount ~= math.floor(amount) then
    return redis.error_reply('amount must be a positive integer')
end

local raw = redis.call('GET', KEYS[1])
local stock = tonumber(raw or '0')
if not stock or stock < 0 or stock ~= math.floor(stock) then
    return redis.error_reply('stock must be a non-negative integer')
end

if stock < amount then
    return -1
end

return redis.call('DECRBY', KEYS[1], ARGV[1])
```

客户端应传入规范的整数字符串，并限制在业务允许的范围内。Redis 内置 Lua 的数字通常采用双精度浮点表示，不能依赖 `tonumber` 精确处理任意 64 位整数。

脚本成功扣减时返回剩余库存，库存不足时返回 `-1`。读取、判断和修改在一次脚本执行中完成，其他客户端不能在检查库存与扣减之间插入命令。

## 3. 与事务的区别

| 能力 | MULTI/EXEC | Lua 脚本 |
| --- | --- | --- |
| 根据中间结果决定后续操作 | 命令先入队，客户端在 `EXEC` 后才获得结果。 | 可以在脚本中读取结果并进行分支判断。 |
| 执行期间防止其他客户端命令插入 | 支持。 | 支持。 |
| 执行出错后自动回滚 | 不支持。 | 不支持。 |

脚本的执行隔离不等于执行结果已同步到磁盘，持久性仍取决于 Redis 的持久化方式与同步策略。

## 4. 错误与回滚

Redis 不会因脚本后续报错而撤销之前成功的修改，例如：

```lua
redis.call('SET', KEYS[1], '1')
redis.call('LPUSH', KEYS[1], 'x')
```

第二条命令因键是字符串而报错，脚本中止，但第一条命令的写入仍然保留。

可以使用 `redis.pcall()` 捕获命令错误，并自行执行补偿操作，例如恢复事先保存的旧值。但 Redis 没有通用的手动回滚命令；补偿还要处理键原本不存在、过期时间、数据类型和补偿本身失败等情况，不能提供数据库回滚那样的保证。

通常应先完成参数、类型和业务条件检查，再执行写操作，并尽量减少写入后仍可能失败的步骤。

## 5. 执行开销

脚本执行期间会阻塞其他客户端的普通命令，应避免长循环或遍历大量数据。脚本超时不代表此前的写入会自动撤销；已经执行写操作的脚本也不能通过 `SCRIPT KILL` 安全终止并回滚。
