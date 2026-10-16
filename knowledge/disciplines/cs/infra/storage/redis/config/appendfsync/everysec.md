---
color: "#5B6CDA"
---

# everysec

`everysec` 是 `appendfsync` 的默认值，通常每秒将 AOF 同步到磁盘一次；需要开启 AOF 才会生效。

## 1. 同步行为

Redis 将 AOF 写入操作系统页缓存，并通常每秒通过后台线程执行一次同步。客户端收到写操作成功的响应时，对应数据可能尚未同步到磁盘。

## 2. 性能与持久性

相比每批写入都同步，`everysec` 减少同步开销，是性能与持久性的常用折中。

故障时通常可能丢失约 1 秒的数据。磁盘延迟、操作系统和硬件故障等因素会影响实际窗口，1 秒并不是严格的数据丢失上限。

## 3. 配置方式

在已开启 AOF 的实例中设置：

```conf
appendfsync everysec
```

也可以动态修改：

```redis
CONFIG SET appendfsync everysec
```

运行时修改不会自动写回配置文件。
