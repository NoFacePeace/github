# Bubblewrap

Bubblewrap（命令名 `bwrap`）是 Linux 沙箱构造工具，通过 namespace 和挂载选项组织受限环境。它不提供统一的默认安全策略，保护范围取决于调用方传入的参数。

## 1. 工作方式

从独立的挂载 namespace 开始，按参数暴露文件、设置读写权限，并选择进程、网络等隔离。无需启动完整 Docker 环境，但仍与宿主共享内核。

## 2. 常见参数

- [参数](parameters/Parameters.md)：文件映射、namespace、会话和系统调用过滤等参数。

## 3. 示例

下面展示只读映射宿主根目录、隔离网络并提供临时目录：

```sh
bwrap \
  --ro-bind / / \
  --unshare-net \
  --unshare-pid \
  --proc /proc \
  --dev /dev \
  --tmpfs /tmp \
  --new-session \
  /bin/sh -c 'printf "sandbox started\\n"'
```

该例仍暴露宿主可读数据，不适合直接运行不可信程序；实际应用应仅映射必要目录。可写映射可能直接修改真实文件。

## 4. 边界与配套机制

- 网络 namespace 不等于域名允许列表。
- CPU、内存配额通常另用 cgroup 等机制。
- 系统调用过滤可配合 seccomp。
- namespace 能否创建取决于内核与宿主策略。

## 5. 参考资料

- [Bubblewrap 官方仓库与使用说明](https://github.com/containers/bubblewrap)
- [命令选项手册源码](https://github.com/containers/bubblewrap/blob/main/bwrap.xml)
- 本机 Linux 手册：`man bwrap`。

资料核对日期：2026-10-22。当前环境为 macOS，示例未在 Linux 上执行验证。
