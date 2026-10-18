# M 创建

M 是 Go 运行时对操作系统线程的表示。运行时根据可运行工作、P 的可用情况及专用线程需求，复用已有 M 或创建新 M。以下以 Go 1.26.3 的主要运行时路径说明，具体实现随平台、构建方式和版本变化。

## 1. 启动时有多少 M

普通可执行程序最初从启动线程对应的 `m0` 开始。运行时在支持的平台上还会创建执行 sysmon 的专用 M；其他线程会根据运行时工作、cgo 和平台需求建立，因此不存在所有 Go 程序统一的“启动完成时 M 数量”。

`GOMAXPROCS` 决定 P 的数量，不决定 M 的数量。初始化多个 P 不代表立即为每个 P 创建一个工作线程。

## 2. 什么时候需要工作 M

| 场景 | 发起处理的执行者与路径 |
| --- | --- |
| 创建或唤醒 G，出现可运行工作 | 当前 M 执行相应运行时路径，按条件调用 `wakep` 等逻辑 |
| sysmon 收回系统调用中的 P | 执行 sysmon 的 M 通过 `retake`、`handoffp` 安排交接 |
| 绑定线程的 G 挂起 | 原 M 在调度路径中通过 `stoplockedm` 交出 P，等待绑定的 G 恢复 |
| 需要专用运行时线程 | 相应初始化或运行时路径直接创建，例如 sysmon |

并不是每次创建 G 都创建 M。已有足够的 M 执行或寻找工作时，通常无需增加工作线程；需要增加时，也优先考虑复用空闲 M。

## 3. 创建 G 后如何安排 M

以 `go f()` 为例，常见路径为：

```text
当前 M 执行 go f()
    ↓
newproc：创建 G，加入可运行结构
    ↓
满足条件时调用 wakep
    ↓
startm：取得空闲 P，寻找空闲 M
    ├─ 有空闲 M → 设置 nextp，通知它接手 P
    └─ 没有空闲 M → newm，创建新线程接手 P
```

`wakep` 会结合空闲 P、自旋 M 等状态限制唤醒，避免多个线程重复增加工作 M。调用 `wakep` 不保证最终创建线程，创建 G 的路径也不保证每次都调用它。

## 4. 真正创建线程的过程

`startm` 在需要新工作 M 时调用 `newm`。专用线程也可以直接通过 `newm` 创建，不必经过普通工作 M 的唤醒路径。

```text
newm
    → allocm：分配 M，准备 g0 等运行时状态
    → 设置启动函数和待接收的 P
    → newm1：进入平台或 cgo 线程创建路径
    → 新线程进入 mstart 相关启动逻辑
    → 执行启动函数，或取得 P 后进入调度
```

Linux 非 cgo 的常见路径通过 `newosproc` 和底层 `clone` 创建线程。cgo、特殊线程状态和其他平台可能采用不同创建路径，不能把所有 `newm` 调用都理解为无条件直接调用 `clone`。

普通工作 M 需要取得 P 才能执行普通 G。sysmon 的专用 M 则不绑定 P，在系统栈上执行受限制的监控逻辑。

## 5. 如何复用与唤醒空闲 M

没有工作、需要停放的普通 M 可以进入空闲 M 列表，并在自己的通知对象 `m.park` 上等待。其他 M 通过 `startm` 选择目标并唤醒：

```text
发起方：取出空闲 M → 设置 M.nextp → notewakeup(&M.park)
                                             ↓
目标 M：从 notesleep(&M.park) 返回 → 取得 P → 继续调度
```

唤醒只是让 OS 线程可以重新被操作系统调度，不保证立即获得 CPU。没有合适的空闲 M 时，运行时才可能创建新线程。

Linux 上，常见 note 实现通过 futex 协调等待和唤醒。调用方使用共享内存地址建立等待关系，不需要把目标线程 ID 传给 futex；具体等待设施依平台和运行时实现而定。

## 6. P 交接由谁执行

### 6.1 系统调用中的 P 被收回

如果 sysmon 收回 P，发起方是执行 sysmon 的 M：

```text
sysmon 的 M → retake → handoffp
                           ├─ 按需 startm：复用或创建工作 M
                           └─ 暂不需要工作 M：P 进入空闲池
```

sysmon 不会拿这个 P 执行业务 G；原 M 的系统调用仍可在内核中继续执行。

### 6.2 锁定线程的 G 挂起

G 通过 `LockOSThread` 绑定 M 后，如果 G 挂起，该 M 不能改去执行无关的普通 G。它通过 `stoplockedm` 主动交出 P，并停放等待；无需 sysmon 事后发起交接。

绑定 G 再次被选中时，调度器通过 `startlockedm` 将执行资源交给其绑定的 M，并通知该 M 恢复执行。这属于线程绑定的专用交接路径，不能直接当作普通空闲 M 的选择过程。

## 7. 源码入口

- `src/runtime/proc.go`：`newproc`、`wakep`、`startm`、`allocm`、`newm`、`newm1`、`stopm`、`handoffp`、`stoplockedm` 和 `startlockedm`。
- `src/runtime/runtime2.go`：M 的 `g0`、`p`、`nextp`、`park` 等字段。
- `src/runtime/os_linux.go`：Linux 上的 `newosproc`。
- `src/runtime/lock_*.go`：note 等待和通知的不同实现。
