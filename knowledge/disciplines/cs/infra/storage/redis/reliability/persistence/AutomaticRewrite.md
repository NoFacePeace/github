---
color: "#5B6CDA"
---

# Automatic Rewrite

## 1. 触发方式

Redis 在 AOF 大小同时达到最小阈值和增长比例阈值时触发自动重写，默认配置通常为：

```conf
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

表示 AOF 至少达到 64 MB，且相对上次重写后的大小增长约 100%（达到约两倍）。首次重写前，使用启动时的 AOF 大小作为比较基准。实际部署可能覆盖这些默认配置。

将 `auto-aof-rewrite-percentage` 设为 `0` 可以关闭自动重写，不影响手动重写。

重写不是按固定时间间隔执行，也不会因为某次删除操作立即触发。已有后台持久化任务等情况可能使重写推迟。
