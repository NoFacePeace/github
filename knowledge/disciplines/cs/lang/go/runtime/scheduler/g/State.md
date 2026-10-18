# G 状态

Go runtime 通过 `g.atomicstatus` 等状态记录 goroutine 的执行阶段，并协调调度、等待、抢占和栈扫描。状态常量主要定义于 `src/runtime/runtime2.go`，具体实现以所用 Go 版本为准。

## 1. 主要状态

| 状态 | 含义 |
| --- | --- |
| `_Gidle` | 刚分配、尚未完成初始化的 G |
| `_Grunnable` | 可以执行，等待调度 |
| `_Grunning` | 正在执行，关联执行它的 M 和 P |
| `_Gsyscall` | 处于系统调用路径，具有相应栈与线程管理约束 |
| `_Gwaiting` | 已暂停执行，由具体等待或运行时协议管理 |
| `_Gdead` | 当前没有活跃的 goroutine 执行，结构可能被复用 |
| `_Gcopystack` | 正在复制或调整栈，需协调其他访问者 |
| `_Gpreempted` | 已响应抢占停止请求，等待暂停请求方接管 |

这不是全部内部状态与组合。`_Grunnable` 不保证 G 必然位于 P 的普通队列，也可能位于全局队列、`runnext` 或调度交接路径中。

## 2. 常见转换

```text
创建：   _Gdead → _Grunnable → _Grunning

等待：   _Grunning → _Gwaiting → _Grunnable

让出：   _Grunning → _Grunnable

结束：   _Grunning → _Gdead

抢占停止：_Grunning → _Gpreempted
                         ↓ suspendG 接管
                     _Gwaiting + 扫描保护
                         ↓ resumeG 按暂停凭据恢复
                     _Grunnable
```

图中省略部分中间状态与同步操作。`resumeG` 不会无条件唤醒原本就在等待的 G；最后一条路径描述的是由本次停止请求暂停正在运行的 G 的典型情况。

## 3. waiting 与 preempted 的区别

### 3.1 已建立的等待协议

`_Gwaiting` 表示 G 已经进入等待，由具体机制决定何时恢复。例如 channel、锁、信号量等待，以及运行时控制下的暂停。

等待记录不一定放在队列中：channel 可以通过 `sudog` 等待队列关联 G，GC 标记 worker 可以通过可用池的节点关联 G，其他机制也可以通过专门变量保存引用。

### 3.2 停止完成的交接状态

`_Gpreempted` 表示目标已响应停止请求，但尚待 `suspendG` 取得暂停控制权。它通常不在普通可运行队列中，也没有专门的停止队列。

目标的 `g` 对象、栈和执行现场仍保留着，暂停请求方持有其指针，运行时也有全局 G 集合等引用。不是只有调用方知道目标存在。

### 3.3 不只是名称不同

普通就绪路径不能直接将 `_Gpreempted` 当成 `_Gwaiting` 处理。运行时需要按状态转换协议接管停止结果、取得扫描权限，再按暂停结果恢复。

因此，两者既有语义区别，也有同步协议区别，不能仅按“有没有保存位置”或“谁写了状态”分类。

## 4. 状态由谁设置

典型等待路径中，G 调用 `gopark`，切到 `g0` 后由 `park_m` 设置 `_Gwaiting`。

抢占停止路径中，其他执行方先设置目标的 `preemptStop` 等请求标志。目标响应后，切到 `g0`，由 `preemptPark` 提交 `_Gpreempted`；随后暂停请求方通过 `suspendG` 接管。

请求标志不等于实际调度状态。设置停止请求时，目标仍可能处于 `_Grunning`，直到它安全响应请求。

## 5. 扫描保护与等待原因

`_Gscan` 是与部分基础状态组合使用的扫描保护位，例如 `_Gscanwaiting`，用于协调栈检查等操作的所有权。它不是“G 正在执行 GC”的同义词。

`waitreason` 则进一步记录等待原因，例如 channel 发送、接收或 GC worker 等待。同样处于 `_Gwaiting` 的 G，可以由完全不同的机制管理。

运行时通过原子状态转换、保护位和相关锁共同维护一致性，不能仅修改一个状态字段就认为完成了挂起或唤醒。

## 6. 源码入口

- `src/runtime/runtime2.go`：G 状态常量及等待原因。
- `src/runtime/proc.go`：调度、挂起、就绪和状态转换辅助函数。
- `src/runtime/preempt.go`：`suspendG`、`resumeG` 及暂停控制。
