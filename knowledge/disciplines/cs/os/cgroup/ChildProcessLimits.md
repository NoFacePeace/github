# Child Process Limits

通过 systemd 启动主进程，由主进程管理委派的 cgroup 子树，对子进程及其后代设置资源限制，主进程不额外设置这些限制。本文使用 Linux cgroup v2，假设其挂载在 `/sys/fs/cgroup`，并使用 systemd 254 或更新版本提供的 `DelegateSubgroup`。

## 1. 目标结构

```text
/sys/fs/cgroup/system.slice/my-app.service/
├─ main/                  主进程，不额外设置 CPU、内存限制
├─ worker-1/              子进程 1 及其后代，共享 1 核 CPU、512 MiB 内存
└─ worker-2/              子进程 2 及其后代，另有自己的额度
```

每个 worker 子组的额度由组内所有进程共享，不是每个进程分别获得一份额度。主进程仍受祖先 cgroup 和整台机器的资源条件约束。

如果直接在 `my-app.service` 上配置 `CPUQuota`、`MemoryMax`，限制会覆盖整个服务，包括主进程，因此需要把主进程与 worker 分开。

## 2. systemd 委派配置

系统服务配置文件 `/etc/systemd/system/my-app.service` 示例：

```ini
[Unit]
Description=Application with child process resource limits

[Service]
User=myuser
ExecStart=/absolute/path/to/your-program
Delegate=cpu memory
DelegateSubgroup=main

[Install]
WantedBy=multi-user.target
```

- `User`：使用实际存在的服务账号，示例中的路径也需要替换。
- `Delegate`：将 CPU、内存控制器对应的子树管理权委派给服务。
- `DelegateSubgroup`：让主进程运行在 `main` 子组，使服务根组本身不驻留进程。

cgroup v2 的 CPU、内存等控制器通常要求向下分配资源的非根组本身不驻留进程，因此不能让主进程留在服务根组，同时又在该组下启用这些控制器。

保存配置后启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl start my-app.service
```

systemd 负责启动和委派；程序仍需实现下面的子组创建与进程迁移逻辑。旧版本 systemd 需要由程序等另行安排主进程迁入子组，再启用控制器，不能直接依赖 `DelegateSubgroup`。

## 3. 确定服务 cgroup 路径

服务启动后，可以查询其 cgroup 路径：

```bash
systemctl show my-app.service --property=ControlGroup --value
```

典型返回值为：

```text
/system.slice/my-app.service
```

加上挂载点后得到完整路径：

```bash
SERVICE_CGROUP="/sys/fs/cgroup$(systemctl show my-app.service \
  --property=ControlGroup --value)"
```

`SERVICE_CGROUP` 是示例中自行定义的 Shell 变量，不是系统内置变量。使用前应确认查询成功且返回值非空，避免误操作挂载根目录。用户级服务应使用 `systemctl --user show`。

程序内部也可读取 `/proc/self/cgroup` 确定自身所在组。在上述配置中，主进程位于服务根组下的 `main`，需要使用其父组作为管理子树的根；不要把主进程所在的 `main` 当成服务根组。路径解析需结合实际挂载点及 cgroup namespace，本文按主机上的常见布局说明。

## 4. 启用控制器并配置子组

以下操作由具备委派权限的主进程完成，Shell 命令用于展示对应的文件操作。

### 4.1 启用控制器

```bash
cat "$SERVICE_CGROUP/cgroup.controllers"
cat "$SERVICE_CGROUP/cgroup.subtree_control"

echo "+cpu +memory" > "$SERVICE_CGROUP/cgroup.subtree_control"
```

`cgroup.controllers` 列出可提供给子组的控制器，`cgroup.subtree_control` 表示已经向子组启用的控制器。已启用时可以跳过写入。

启用控制器本身不设置任何资源上限，只是允许下层子组配置相应限制。

### 4.2 创建子组并设置限制

```bash
mkdir "$SERVICE_CGROUP/worker-1"

# 每 100000 微秒周期最多使用 100000 微秒 CPU 时间
echo "100000 100000" > "$SERVICE_CGROUP/worker-1/cpu.max"

# 内存硬限制为 512 MiB
echo 536870912 > "$SERVICE_CGROUP/worker-1/memory.max"
```

CPU 配额耗尽通常导致限流；内存达到硬限制且无法回收时，可能触发组内 OOM。交换空间有独立的控制项，`memory.max` 不代表内存与交换空间的总上限。

## 5. 创建子进程并迁移

### 5.1 子进程自行加入

创建进程和执行实际任务是两个步骤。`fork()` 返回后，子进程已经存在并拥有 PID，可以先加入目标 cgroup，再通过 `exec()` 执行任务。

```text
主进程：创建并配置 worker-1
  │
  └─ fork()：子进程已产生，拥有 PID，最初继承 main 组
       │
       ├─ 获取自己的 PID
       ├─ 将 PID 写入 worker-1/cgroup.procs
       ├─ 写入失败则退出
       └─ 写入成功后 exec() 执行实际任务
```

伪代码：

```c
// 主进程已创建并配置 worker-1
pid = fork();

if (pid == 0) {
    // 子进程准备阶段，还没有执行实际任务
    if (write_pid_to_cgroup(worker_procs_path, getpid()) != SUCCESS) {
        _exit(1);
    }
    exec_worker();
    _exit(1); // exec 失败
} else if (pid < 0) {
    handle_fork_error();
}
```

这里的辅助函数只表示流程，并非可直接编译的实现。多线程程序在 `fork()` 后、`exec()` 前还有调用安全约束，实现时应采用适合运行时的进程启动方式。

### 5.2 权限要求

子进程不会因为是主进程创建的就自动获得 cgroup 管理权限。它能否迁移，取决于委派范围和运行身份。

在上述配置中，systemd 将相关管理权限委派给 `myuser`。子进程仍以该账号运行时，通常可以在委派子树内迁移自己。cgroup v2 迁移主要要求能写目标组的 `cgroup.procs`，以及源组与目标组共同祖先的 `cgroup.procs`。

从 `main` 迁入 `worker-1` 时，共同祖先就是 `my-app.service` 对应的组。如果子进程需要切换到其他账号，可以先加入目标组，再降低权限。

也可以由主进程写入子进程 PID，但必须先让子进程等待，迁移成功后再通知它执行任务，避免实际任务先运行再受到限制。

## 6. 后代进程自动继承

子进程加入 `worker-1` 后，再创建的后代进程会自动继承这个组，无须再次逐个写入 PID。

```text
worker-1/                    合计：1 核 CPU、512 MiB 内存
└─ 子进程
   ├─ 孙进程 A
   └─ 孙进程 B
      └─ 更下一代进程
```

加入目标组之前已经创建的后代，不会自动跟随迁移，需要单独处理。`exec()` 替换程序时不会改变 PID 或 cgroup 归属。

如果后代进程仍持有委派子树的管理权限，它也可能主动迁出受限组。这种方案适用于配合管理的程序；要约束不可信任务，还需收紧其身份与 cgroup 写入权限。

## 7. 检查与清理

```bash
cat "$SERVICE_CGROUP/worker-1/cgroup.procs"
cat "$SERVICE_CGROUP/worker-1/cpu.stat"
cat "$SERVICE_CGROUP/worker-1/memory.current"
cat "$SERVICE_CGROUP/worker-1/memory.events"
```

可通过 `/proc/<PID>/cgroup` 检查具体进程的归属。`cgroup.procs` 只列出该组直接包含的进程，不递归列出下层子组中的进程。

主进程应负责等待并回收直接子进程；确认组内全部进程退出、下层子组也已删除后，再用 `rmdir` 删除 worker 组。直接子进程退出不代表它创建的后代已全部退出。
