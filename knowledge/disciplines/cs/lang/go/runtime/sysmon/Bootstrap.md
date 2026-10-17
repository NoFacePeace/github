# Bootstrap

sysmon 由 `runtime.main` 创建专用 M，在新 M 的 g0 上直接进入监控循环，始终不绑定 P。它的启动不经过普通 G 的运行队列。

## 1. 在程序启动中的位置

以普通 Go 可执行程序的 amd64 启动路径为例：

```text
runtime.rt0_go
  → 建立 m0、g0
  → schedinit()：初始化调度器，创建 P，将 P0 绑定给 m0
  → newproc(runtime.main)：创建主 G，放入 P0 的本地可运行结构
  → mstart → mstart0 → mstart1 → schedule
  → 执行主 G：runtime.main
      → 启动 sysmon
      → 执行 runtime 初始化任务，启用 GC 等
      → 执行用户包 init
      → main.main()
```

最初由 m0 在 g0 上完成初始化；主 G 被调度后，才执行 `runtime.main`。sysmon 在包初始化之前启动，因此用户包的 `init()` 阶段也能受到监控。

## 2. 创建专用 M

`runtime.main` 中的启动代码：

```go
mainStarted = true
if haveSysmon {
    systemstack(func() {
        newm(sysmon, nil, -1)
    })
}
```

`systemstack` 将创建 M 的操作切到当前 m0 的 g0 上执行；真正的 sysmon 随后在新 M 的 g0 上运行。当前 wasm 没有 sysmon 线程支持，由 `haveSysmon` 控制是否创建。

| 参数 | 含义 |
| --- | --- |
| `sysmon` | 新 M 的启动函数，保存到 `m.mstartfn` |
| `nil` | 不为该 M 分配 P |
| `-1` | 没有预分配 M ID，由运行时分配 |

```text
runtime.main（主 G）
  → systemstack（m0.g0）
  → newm(sysmon, nil, -1)
      → allocm：分配 M 和 g0，设置 m.mstartfn
      → newm1：通过 OS 或 cgo 线程创建路径启动线程
          → mstart → mstart0 → mstart1（新 M 的 g0）
              → m.mstartfn()
                  → sysmon()
```

创建线程后，主 G 和 sysmon 可以并行推进；这里没有等待 sysmon 就绪的握手。Linux 非 cgo 路径通过 `newosproc → clone` 创建线程，其他平台使用各自的线程创建实现。

## 3. 为什么不需要 P

`mstart1` 中，启动函数在获取 P 之前调用：

```go
if fn := gp.m.mstartfn; fn != nil {
    fn()
}
if gp.m != &m0 {
    acquirep(gp.m.nextp.ptr())
    gp.m.nextp = 0
}
schedule()
```

`sysmon()` 进入无限循环，不返回，因此该 M 不会执行后面的 `acquirep` 和 `schedule`。

P 提供执行普通 Go 代码所需的运行时资源，不是 OS 线程执行机器指令的前提。sysmon 执行受限制的运行时代码，入口标记 `//go:nowritebarrierrec`，不能像普通 G 一样任意分配 Go 对象、执行写屏障。

这样，即使所有 P 都忙于执行 G 或处于系统调用状态，sysmon 仍能获得 OS 调度，检查运行时进展。只有一个逻辑 CPU 时，sysmon 的 M 与业务 M 由操作系统轮流调度。

## 4. 进入监控循环

`sysmon()` 先在调度锁保护下执行：

```go
sched.nmsys++
checkdead()
```

将自身计入系统 M，使死锁检测不会把它当作可执行普通 G 的工作线程。随后初始化监控状态，进入休眠、检查和唤醒工作的循环。

首轮先休眠约 `20µs`，后续根据活动情况调整间隔。具体职责、休眠节奏与抢占检查见 [Sysmon](Sysmon.md)。

## 5. 源码入口

以下调用链已对照本地 Go 源码 `master` 的 `714d9afb66` 核对；线程创建细节随平台和构建方式变化。

- `src/runtime/asm_amd64.s`：`rt0_go`、`mstart`。
- `src/runtime/proc.go`：`main`、`allocm`、`newm`、`newm1`、`mstart0`、`mstart1`、`sysmon`。
- `src/runtime/os_linux.go`：Linux 的 `newosproc`。
