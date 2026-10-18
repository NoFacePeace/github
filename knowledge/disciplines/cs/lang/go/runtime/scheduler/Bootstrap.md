# 调度器启动

Go 调度器的启动包括建立初始 M 与系统栈、初始化调度资源、创建主 G，以及首次选择并执行 G。以下以普通 Go 可执行程序的常见 amd64 启动路径为例，具体入口随平台、构建方式和 Go 版本变化。

## 1. 整体流程

```text
平台启动入口 → runtime.rt0_go
    ↓
建立 m0、g0 及线程局部运行时上下文
    ↓
schedinit：初始化运行时和调度资源，准备 P
    ↓
newproc：创建执行 runtime.main 的主 G
    ↓
mstart → mstart0 → mstart1
    ↓
schedule → execute
    ↓ 切换到主 G 栈
runtime.main → 初始化任务 → main.main
```

调度器先建立执行普通 G 所需的基础设施，之后主 G 才能被调度。用户包的初始化函数不是在最初的启动汇编阶段执行。

## 2. 建立 m0 与 g0

启动线程及其初始化职责参见 [m0](m/m0.md)，系统 G 与系统栈的用途参见 [g0](g/g0.md)。

启动代码准备初始 `g0` 的栈边界、栈检查信息及获取当前 G 所需的线程上下文，并关联全局 `m0` 与 `g0`。

`m0` 对应启动线程，`g0` 提供该线程执行运行时代码的系统栈。此时还没有开始执行主 G 的业务调用栈；早期初始化也不能假设完整的调度与内存管理设施已经可用。

## 3. 初始化调度资源

`schedinit` 协调运行时初始化，包括线程、内存管理、模块信息及调度器相关准备。它不是只初始化一个运行队列的函数。

调度器根据启动配置确定 P 的数量，通过 `procresize` 等逻辑创建或配置 P，并使初始 M 获得执行普通 Go 代码所需的 P。各 P 的运行队列等资源随之建立。

P 的数量受 `GOMAXPROCS` 及对应版本默认策略影响，不等于启动时立即创建同样数量的操作系统线程。M 会按运行时需要进一步创建或唤醒。

## 4. 创建主 G 并首次调度

启动路径通过 `newproc` 创建执行 `runtime.main` 的主 G，将其置为可运行并加入可运行结构。`newproc` 不在调用点直接切到主 G 执行。

随后初始 M 进入 `mstart` 相关路径，在 `g0` 栈上完成必要准备并调用 `schedule`。调度器选中主 G 后，通过执行路径建立 G、M 的运行关系并切换到主 G 的栈。

```text
m0 的 g0 栈：初始化 → 创建主 G → 首次调度
                                      ↓
主 G 栈：                        runtime.main
```

`runtime.main` 与用户的 `main.main` 不同，前者负责运行时进入 Go 程序后的启动协调，后者是用户程序入口。

## 5. runtime.main 的后续准备

主 G 执行后，`runtime.main` 按运行时协议完成后续准备，例如在支持的平台启动 sysmon、执行运行时初始化任务、启用 GC，并推进用户包初始化，最后调用 `main.main`。

部分后台 G、专用 M 和 GC 资源在后续初始化或首次需要时才建立，不是全部在 `schedinit` 一次性创建。sysmon 也有独立的启动和执行路径，不通过普通业务 G 队列执行监控循环。

`main.main` 返回后，运行时结束进程，不会等待所有其他 goroutine 自行退出。

## 6. 源码入口

- `src/runtime/rt0_*.s`：平台相关进程入口。
- `src/runtime/asm_*.s`：`rt0_go`、`mstart` 等启动与执行切换入口。
- `src/runtime/proc.go`：`schedinit`、`procresize`、`newproc`、`mstart0`、`mstart1`、`schedule` 和 `runtime.main`。
