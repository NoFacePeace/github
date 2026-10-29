# Parameters

Bubblewrap 通过命令行参数构造沙箱环境，基本形式为：

```sh
bwrap [选项...] 要执行的命令 [命令参数...]
```

## 1. 文件系统

| 参数 | 作用 |
| --- | --- |
| `--ro-bind SRC DEST` | 将宿主路径映射到沙箱内的目标路径，并以只读方式挂载。 |
| `--bind SRC DEST` | 可写映射路径，实际访问仍受文件权限约束。 |
| `--tmpfs DEST` | 在目标路径挂载临时文件系统。 |
| `--proc DEST` | 挂载 proc 文件系统，常用目标为 `/proc`。 |
| `--dev DEST` | 提供基础设备环境，常用目标为 `/dev`。 |

映射宿主根目录会暴露宿主可读数据；可写映射可能直接修改宿主文件。实际沙箱需要按任务选择映射范围。

## 2. Namespace 与会话

| 参数 | 作用 |
| --- | --- |
| `--unshare-net` | 创建独立网络 namespace，不自动配置对外连接。 |
| `--unshare-pid` | 创建独立 PID namespace，通常配合 `--proc /proc`。 |
| `--new-session` | 创建新会话，使任务与原会话分离。 |

Namespace 能否创建取决于内核支持、权限和宿主策略。

## 3. 系统调用过滤

- [--seccomp FD](Seccomp.md)：从已打开的文件描述符读取 BPF 过滤指令，在启动任务前安装 seccomp 过滤器。

## 4. 查看本机选项

```sh
bwrap --help
man bwrap
```

不同版本可能提供不同选项，应以实际安装版本的帮助和手册为准。
