# 准备启动文件与资源限制

读取执行资源限制，准备任务目录与 manifest，生成 bwrap 需要的 hosts 和 seccomp 辅助文件。

## 1. 文件准备

任务目录包括 workspace、config、session、memory、home 和 tmp。manifest 保存执行描述，不放入完整敏感输入。

hosts 与 seccomp 等私有辅助文件放在任务挂载范围外的进程管理目录，并设置受限权限；普通执行不能直接访问这些宿主管理文件。

## 2. 资源限制

准备适用的任务 cgroup 根目录，再为本次执行创建 cgroup，应用执行资源限制。cgroup 管理进程资源，与 Redis 中预留 Runner 并发名额的容量集合不是同一机制。

## 3. 后续步骤

进入[准备 MCP 与 Git 环境](PrepareToolEnvironment.md)。
