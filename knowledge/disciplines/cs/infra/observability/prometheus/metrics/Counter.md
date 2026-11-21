# Counter

Counter（计数器）记录持续累加的值，例如请求总数、超时总数或发送的字节数。值可以增加或保持不变，也可能在进程重启时归零。

## 1. 查询增长速率

如果库存查询每超时一次，就把 `inventory_query_timeouts_total` 加 1，可以用 PromQL 的 `rate` 查看最近 5 分钟平均每秒超时多少次。计算时会考虑计数器重置。

```promql
rate(inventory_query_timeouts_total[5m])
```
