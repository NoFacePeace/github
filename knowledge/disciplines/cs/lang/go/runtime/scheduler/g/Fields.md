# G 字段

`g` 是 goroutine 的运行时表示，定义于 `src/runtime/runtime2.go`。以下整理理解栈、调度与等待时常用的字段，并非完整结构定义，具体字段以所用 Go 版本为准。

## 1. 栈与恢复现场

| 字段 | 作用 |
| --- | --- |
| `stack` | 记录栈的地址范围 |
| `stackguard0` | 用于普通栈检查，也参与协作式抢占 |
| `sched` | 保存恢复执行所需的上下文 |

## 2. 调度状态与执行关系

| 字段 | 作用 |
| --- | --- |
| `atomicstatus` | 记录 G 的基础状态及相关保护状态 |
| [m](m.md) | 关联执行该 G 的 M，停止执行时可解除绑定 |
| `goid` | 运行时内部的 goroutine 标识 |
| `preempt` | 抢占请求标志 |
| `preemptStop` | 请求目标响应抢占后停止，供暂停方接管 |

请求标志与实际调度状态不同：提出抢占请求不代表 G 当场已停止。`goid` 也不是应用可依赖的稳定公开接口。

## 3. 等待记录

| 字段 | 作用 |
| --- | --- |
| `waitreason` | 记录等待原因 |
| [waiting](waiting.md) | 指向相关 `sudog` 等待记录链，节点通过 `waitlink` 连接 |

`waitreason` 描述等待原因，`waiting` 关联具体等待记录。并非所有处于 `_Gwaiting` 的 G 都必须有非空的 `waiting` 链。
