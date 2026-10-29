# Seccomp

seccomp（Secure Computing）是 Linux 内核限制进程可执行哪些系统调用的机制，常用于容器、沙箱和服务加固，减少程序能够请求的内核操作。

## 1. 基本概念

程序读写文件、创建进程、操作网络时，通常需要通过系统调用请求内核完成操作。

| 系统调用 | 作用 |
| --- | --- |
| `read` / `write` | 读写文件描述符。 |
| `openat` | 打开文件。 |
| `execve` | 执行程序。 |
| `ptrace` | 跟踪或调试进程。 |
| `mount` | 挂载文件系统。 |

seccomp 可以允许调用、拒绝并返回指定错误，或者终止线程或进程等。过滤器模式还支持通知用户态监控程序，但需要额外实现通知处理机制。

例如，一个普通计算任务通常不需要挂载文件系统或调试其他进程，可以禁止相应的系统调用。过滤器允许调用并不代表调用必然成功，文件权限、capability 等其他检查仍然有效。

## 2. 模式与规则边界

| 模式 | 特点 |
| --- | --- |
| 严格模式 | 只允许 `read`、`write`、`_exit` 和信号返回等极少数调用，适用范围有限。 |
| 过滤器模式（seccomp-BPF） | 通过 BPF 过滤器，根据架构、系统调用编号和参数值等决定处理动作，是容器常用的方式。 |

过滤器可以检查整数参数值，但普通 seccomp-BPF 不能直接解引用用户态指针，因此不能直接读取 `openat` 的路径字符串，表达“只允许访问某个目录”。路径访问控制通常需要文件权限、Landlock、SELinux 或 AppArmor 等机制配合。

过滤规则可以采用两种思路：

- 默认允许，拒绝部分已知不需要的调用，较易兼容，但覆盖范围有限。
- 默认拒绝，只允许任务需要的调用，限制更严格，但需要覆盖动态链接器、运行时和异常处理等路径。

## 3. 通过 systemd 使用

### 3.1 服务配置

在支持 seccomp 的 Linux 和 systemd 环境中，可以在服务配置中加入：

```ini
[Service]
ExecStart=/absolute/path/to/your-program
NoNewPrivileges=yes
SystemCallFilter=~@mount @reboot @swap @debug
SystemCallErrorNumber=EPERM
```

- `SystemCallFilter` 开头的 `~` 表示拒绝列出的调用或调用集合。
- `@mount`、`@reboot`、`@swap`、`@debug` 是 systemd 定义的调用集合。
- `SystemCallErrorNumber=EPERM` 让被过滤的调用返回“操作不允许”。
- `NoNewPrivileges=yes` 禁止进程及其后代通过 `exec` 时的 setuid、文件 capability 等机制获得额外权限；它不会撤销进程已有的权限。

这是一份演示用拒绝列表，实际配置需要根据服务所需的操作调整。过滤器作用于服务主进程及其后续子进程，不适用于“主进程不限制，仅限制某个子进程”的需求。

### 3.2 应用配置与检查

```bash
sudo systemctl daemon-reload
sudo systemctl restart my-app.service

systemctl status my-app.service
journalctl -u my-app.service
```

可以检查调用集合的成员：

```bash
systemd-analyze syscall-filter @mount
```

集合内容可能随 systemd 版本和架构变化，应以实际环境为准。命令示例需要替换实际程序路径和服务名。

## 4. 通过 libseccomp 使用

### 4.1 禁止 uname 的 C 示例

libseccomp 帮助程序构建和加载过滤器。下面默认允许系统调用，仅将 `uname` 拒绝为 `EPERM`：

```c
#include <errno.h>
#include <stdio.h>
#include <sys/utsname.h>
#include <seccomp.h>

int main(void) {
    scmp_filter_ctx ctx = seccomp_init(SCMP_ACT_ALLOW);
    if (ctx == NULL)
        return 1;

    int rc = seccomp_rule_add(ctx, SCMP_ACT_ERRNO(EPERM),
                             SCMP_SYS(uname), 0);
    if (rc < 0) {
        fprintf(stderr, "seccomp_rule_add failed: %d\n", rc);
        seccomp_release(ctx);
        return 1;
    }

    rc = seccomp_load(ctx);
    seccomp_release(ctx);
    if (rc < 0) {
        fprintf(stderr, "seccomp_load failed: %d\n", rc);
        return 1;
    }

    struct utsname info;
    if (uname(&info) == -1) {
        perror("uname");
        return 0;
    }

    fprintf(stderr, "uname unexpectedly succeeded\n");
    return 1;
}
```

Ubuntu/Debian 环境中，保存为 `demo.c` 后安装依赖并编译：

```bash
sudo apt install build-essential libseccomp-dev
gcc -Wall -Wextra demo.c -o demo -lseccomp
./demo
```

典型输出如下，错误文本可能随系统语言变化：

```text
uname: Operation not permitted
```

`seccomp_release` 只释放用户态配置对象，不会解除已经加载到内核的过滤器。

### 4.2 安装权限与多线程

无相应权限的进程安装过滤器前，需要设置 `no_new_privs`。libseccomp 默认会在加载时请求设置它，因此上述示例在内核和系统策略支持时通常不需要 root。外层容器或安全策略仍可能限制安装操作。

过滤器默认安装在调用线程上。多线程程序若要同步现有线程，需要在加载前配置 `SCMP_FLTATR_CTL_TSYNC`，并检查相关操作的返回值。不能仅因一个线程安装成功，就认为其他已有线程也被限制。

## 5. 仅限制子进程

主进程需要保留系统调用能力时，可以让子进程在执行任务前自行安装过滤器：

```text
主进程：不安装该过滤器
  │
  └─ 创建子进程：此时已有 PID
       ├─ 加入目标 cgroup（如果需要资源限制）
       ├─ 完成必要的权限与环境准备
       ├─ 安装 seccomp 过滤器，失败则退出
       └─ exec 执行实际任务
```

子进程的过滤器不会反向影响主进程。安装后再创建的子进程会继承过滤器，通过 `exec` 替换程序时过滤器仍生效。

规则必须允许 `execve` 或实际使用的执行接口，以及程序启动所需的系统调用，否则任务可能在启动阶段失败。多线程程序中还需要遵守 `fork` 后、`exec` 前的调用安全约束，可采用单独的启动辅助程序完成准备。

已安装的过滤器不能卸载或替换为更宽松的规则。后续安装的过滤器会与已有过滤器共同生效，不能借此取消已有拒绝。

## 6. 与 namespace、cgroup 的关系

| 机制 | 主要作用 |
| --- | --- |
| Namespace | 隔离进程看到和共享的系统资源环境。 |
| Cgroup | 限制、分配和统计一组进程的资源。 |
| Seccomp | 过滤进程能够请求的系统调用。 |

三者可以组合使用，但职责不同。seccomp 本身不限制 CPU、内存额度，也不能代替完整的文件访问控制或容器隔离。
