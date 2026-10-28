# 读取执行输入

Executor 通过 manifest 文件和 stdin 两个入口接收本轮执行描述与敏感数据。

## 1. 处理顺序

从[读取 manifest](ReadManifest.md)开始，沿后续链接处理：

```text
解析 run 与 --manifest 参数
  → 读取、解析并校验执行描述
  → 读取敏感输入
  → 校验 executionId 与凭证类型
  → 合成最终执行配置
```

## 2. 存储边界

此阶段读取 Runtime 已准备的本地 manifest 和进程管道，不直接查询 MySQL、Redis 或对象存储。Executor 不再次向 Server 领取同一份执行凭证。
