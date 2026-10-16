---
color: "#5B6CDA"
---

# Manual Rewrite

## 1. 触发方式

在已开启 AOF 的实例中执行：

```redis
BGREWRITEAOF
```

命令请求后台重写，返回成功受理不代表重写已经完成；如果已有重写任务，命令可能返回错误，如果正在执行 RDB 后台保存，重写可能被安排稍后执行。

## 2. 确认结果

使用 `INFO persistence` 查看 `aof_rewrite_in_progress`、`aof_rewrite_scheduled` 和 `aof_last_bgrewrite_status`，确认任务是否正在执行、等待执行或已经成功完成。
