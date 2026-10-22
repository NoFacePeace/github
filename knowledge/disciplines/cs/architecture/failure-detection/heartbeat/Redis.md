# Redis

使用 Redis 实现节点心跳时，可以用带 TTL 的 String 保存节点详情，用 ZSet 维护按过期时间排序的在线索引。节点周期性上报，调度服务查询并筛选可用节点。

## 1. 数据模型

| 数据 | Redis 类型 | 内容 |
| --- | --- | --- |
| 单节点心跳详情 | String | 节点身份、进程实例、状态、能力和资源等 JSON，设置 TTL。 |
| 节点在线索引 | ZSet | member 为节点 ID，score 为心跳过期时间戳。 |

示例键名：

```text
workers:{pool-a}:heartbeat:node-1
workers:{pool-a}:heartbeats
```

两个 key 使用相同的 hash tag `{pool-a}`，便于在 Redis Cluster 中由同一段 Lua 原子操作。

String 中的示例内容：

```json
{
  "nodeId": "node-1",
  "instanceId": "boot-uuid-1",
  "status": "online",
  "expiresAt": 1792630830000,
  "capabilities": ["job-execution"],
  "maxConcurrency": 4,
  "activeJobs": 2
}
```

ZSet 的逻辑内容：

```text
node-1 → 1792630830000
node-2 → 1792630840000
```

String 用于直接查询详情，ZSet 用于枚举仍有效的节点。仅判断一个已知节点是否在线时，可以只使用带 TTL 的 String。

## 2. 上报与续约

例如节点每 10 秒上报一次，每次刷新 30 秒有效期。一次上报同时完成：

1. 检查节点 ID 是否已由另一个进程实例占用。
2. 写入最新心跳详情，刷新 String TTL。
3. 更新 ZSet 中该节点的过期时间分数。

同一 nodeId 在 ZSet 中只有一个成员，重复上报更新 score，不会新增重复节点。

下面的 Lua 展示原子上报的核心流程。调用方通过 `KEYS` 显式传入详情 key 和索引 key，通过 `ARGV` 传入详情 JSON 和有效期毫秒数：

```lua
local incoming = cjson.decode(ARGV[1])
local ttlMs = tonumber(ARGV[2])
if not ttlMs or ttlMs <= 0 or ttlMs ~= math.floor(ttlMs) then
    return redis.error_reply('invalid heartbeat ttl')
end
if type(incoming.nodeId) ~= 'string' or incoming.nodeId == ''
    or type(incoming.instanceId) ~= 'string' or incoming.instanceId == '' then
    return redis.error_reply('invalid heartbeat identity')
end

local old = redis.call('GET', KEYS[1])
if old then
    local existing = cjson.decode(old)
    if existing.instanceId ~= incoming.instanceId then
        return redis.error_reply('node identity conflict')
    end
end

local time = redis.call('TIME')
local nowMs = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)
local expiresAt = nowMs + ttlMs
incoming.expiresAt = expiresAt
local payload = cjson.encode(incoming)

redis.call('SET', KEYS[1], payload, 'PX', ttlMs)
redis.call('ZADD', KEYS[2], expiresAt, incoming.nodeId)
return expiresAt
```

示例假设 key 类型由应用约定并正确维护，生产环境还需限制有效期和字段范围、验证调用身份及 key 与 nodeId 的对应关系。Lua 不提供写入失败后的自动回滚，不能允许其他业务把同一索引 key 改为不兼容类型。

这里由 Redis 时间计算 expiresAt，使 TTL 与索引期限采用一致的时间依据。若由客户端计算期限，则需要控制客户端时钟偏差。

## 3. 查询指定节点

已知节点 ID 时，直接读取详情：

```redis
GET workers:{pool-a}:heartbeat:node-1
```

读取后检查：

- 详情存在，且 expiresAt 未到期。
- 状态允许接收新任务，例如 `online`。
- 能力、版本等满足当前任务要求。

在线只表示近期仍在上报，不保证有空闲名额。派单前应另外原子检查并预留容量，不能仅依赖心跳快照中的 activeJobs。

## 4. 查询在线节点列表

先从 ZSet 清理并筛选 ID，再批量读取 String 详情：

```text
获取当前时间 nowMs
    ↓
ZREMRANGEBYSCORE：删除 score ≤ nowMs 的过期索引
    ↓
ZRANGEBYSCORE：读取 score > nowMs 的节点 ID
    ↓
MGET：读取各节点心跳 JSON
    ↓
过滤缺失、过期、未就绪或能力不匹配的节点
```

命令示例：

```redis
ZREMRANGEBYSCORE workers:{pool-a}:heartbeats -inf 1792630830000
ZRANGEBYSCORE workers:{pool-a}:heartbeats (1792630830000 +inf
MGET workers:{pool-a}:heartbeat:node-1 workers:{pool-a}:heartbeat:node-2
```

`MGET` 的 key 列表由上一条查询返回的 ID 动态构造，无 ID 时直接返回空列表。通过 shell 执行时，需要给含括号的边界参数加引号。

这组查询可以分步执行，不必放进同一段 Lua。因为查询过程中可能发生过期、续约或退出，最终必须校验详情，不能只信任 ZSet 中的索引。查询时间也应与写入时采用一致的时间基准。

## 5. 正常注销与异常过期

正常退出时，通过 Lua 比较 instanceId；确认详情仍属于本次进程后，再执行：

```text
DEL 该节点的心跳详情 key
ZREM 在线索引中的该节点 ID
```

比较与删除必须放在同一个原子操作中，否则旧进程可能误删新进程已更新的心跳。详情不存在或实例不匹配时，不应无条件删除索引。

异常退出时无法注销：

- String 由 TTL 自动过期。
- ZSet 成员不会因 score 到期自动删除，由在线列表查询或定期清理移除。
- 移除在线索引只影响节点发现与调度，不会停止原进程、释放任务绑定或迁移工作目录。

## 6. 实例冲突与重启

稳定的 nodeId 用于表示节点，instanceId 每次进程启动重新生成。详情仍有效且实例不同，可以拒绝新上报，避免两个进程同时冒用同一个节点身份。

这种策略下，新实例可能需要等待旧详情过期，或经过受控交接才能上线。旧详情过期后，旧进程仍可能恢复上报，因此仅比较当前心跳的 instanceId 不是完整的执行隔离方案；如需严格禁止旧实例重新取得权力，应增加注册代次、实例认证或 fencing 校验。

同一实例的心跳上报也应避免重叠，以免较旧的状态在较新状态之后写入；需要并发上报时，可以增加单调递增序号并校验。

## 7. 故障与规模边界

- Redis 请求失败时，不应把它直接解释为所有节点已死亡，应区分协调存储故障与节点失联，并制定暂停派单或降级策略。
- 过期只是失联判断，旧任务可能仍在运行；任务接管需要独立的执行租约、结果去重和失效校验。
- 心跳详情与索引可能受内存淘汰及故障恢复影响，应为协调状态选择合适的容量、淘汰和持久化策略。
- 大量节点不宜每次全量读取，可以按资源池分区、限制查询量并控制清理批次；单次长 Lua 会阻塞其他普通命令。
- 10 秒上报、30 秒过期只是示例，应根据网络抖动、进程暂停和检测时延要求调整，并可给上报周期增加随机抖动以减少同时上报。
