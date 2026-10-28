# Runtime

Server 派发作业后，Runtime 中的 Runner 接收 Job，校验执行条件，准备环境并启动执行器。

## 1. 阶段流程

从[消费 Job](consume-job/ConsumeJob.md)开始，读取作业并交给 Runner 处理。

```text
从 Redis Stream 接收 Job
  → 校验目标节点、范围与租约
  → 检查重复执行及维护状态
  → 接受任务并回传 accepted 事件
  → 准备凭证、资源与隔离环境
  → 启动执行器
  → 持续回传进度与输出
  → 执行收尾、启动后台备份
  → 发布最终结果
```

## 2. 职责边界

Runtime 承载执行环境与进程生命周期；Server 负责任务业务记录和调度，Executor 负责适配具体 Agent 引擎。

Job 已投递不等于 Runner 已接受，接受任务也不等于 Agent 已启动，各阶段需通过对应事件与状态区分。
