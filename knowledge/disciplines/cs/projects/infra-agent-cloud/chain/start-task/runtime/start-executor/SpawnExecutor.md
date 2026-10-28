# 创建 Executor 进程

将沙箱启动命令纳入执行 cgroup，使用 `spawn` 创建独立进程组，捕获标准输出、标准错误及退出事件。

## 1. 加入执行 cgroup

Runner 直接通过 `spawn` 创建的最外层程序是 `/bin/sh`。包装脚本先将自己的 PID 写入本轮执行分组，再用 `exec` 替换为后续网络与沙箱启动程序：

```sh
set -eu
printf '%s\n' "$$" > "$1/cgroup.procs"
shift
exec "$@"
```

`$1` 是受控的 cgroup 目录参数，`$$` 是该 shell 的 PID。写入者不是 Executor，也不是 Runner 将自身移入执行组；Runner 仍留在 supervisor 分组。

后续创建的进程继承 cgroup 归属，`exec` 不改变当前进程的组。Executor、Agent 及其子进程因此受本轮资源限制。按参数传递路径与命令，不将任务提示词拼成 shell 代码。

## 2. 进程登记

操作系统确认进程创建后，将执行标识、PID、身份标记、私有文件清单与启动时间登记到本地进程注册信息中，供停止、清理与重启恢复使用。

这一步主要写 Runner 本地文件，不是向 MySQL 新增进程表记录。

登记的是 `spawn` 返回的直接子进程 PID，在隔离网络路径中可能对应外层 RootlessKit 启动进程，不能直接认定为沙箱内 Executor 或 Agent 引擎的 PID。

## 3. 敏感输入

通过标准输入管道传入执行凭证等敏感数据，避免放入 argv。普通模式写入后关闭输入；支持运行中指令的模式可保持管道打开，供后续控制使用。

取得适用的沙箱子进程 PID，便于后续发送停止信号。进程登记或输入交付失败时，终止适用的进程并清理登记，不能当作正常启动成功。

FD 3 用于取得 bwrap 子进程信息，FD 4 用于加载 seccomp 策略，与传递凭证的 stdin 分别使用。最终停止与收尾还结合进程组和 cgroup，避免只终止一个 PID 而遗留后代进程。

## 4. 后续步骤

返回可控进程句柄，进入[建立执行监控](MonitorExecutor.md)。Executor 随后读取 manifest 和敏感输入，再配置并启动具体 Agent 引擎。
