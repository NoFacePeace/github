# Seccomp

Bubblewrap 的 `--seccomp FD` 参数从已打开的文件描述符读取 seccomp BPF 过滤指令，在启动目标程序前安装过滤器。Bubblewrap 负责读取和提交规则，Linux 内核负责执行系统调用过滤。

## 1. 参数取值

```sh
bwrap --seccomp FD [其他选项...] 要执行的命令
```

`FD` 必须是文件描述符编号，例如 `4`，不能直接填写文件路径：

```text
--seccomp 4           正确：读取 FD 4
--seccomp rules.bpf   错误：这里不接受文件路径
```

文件描述符是进程引用已打开文件、管道等 I/O 对象的整数句柄。通常 `0`、`1`、`2` 分别用于标准输入、标准输出和标准错误，其他可用编号可以用于传递规则。

`4` 没有特殊含义，也不是固定值；需要保证该编号在 Bubblewrap 进程中确实引用可读取的规则数据。

## 2. BPF 二进制文件

seccomp BPF 文件保存的是经典 BPF（cBPF）过滤指令的二进制编码，供内核判断系统调用是否允许。它不是普通 Linux 可执行程序，也不是 eBPF ELF 对象文件或 JSON 规则描述。

例如，“禁止 `uname`，返回 `EPERM`，其他调用允许”的规则，生成的指令会检查架构、系统调用编号，并返回相应处理动作。

程序可通过 libseccomp 构造规则，并调用 `seccomp_export_bpf(ctx, fd)` 将过滤指令导出到文件。TypeScript 也可以组织规则，再通过生成代码或辅助工具转换成兼容的 BPF 指令；仅把规则文本写进文件并不能供此参数使用。

## 3. 通过文件描述符传入

假设已经生成有效的 `rules.bpf`，可以用 Shell 重定向打开文件并将 FD 4 传给 Bubblewrap：

```sh
bwrap \
  --ro-bind / / \
  --seccomp 4 \
  /absolute/path/to/agent \
  4< rules.bpf
```

其中 `4< rules.bpf` 是 Shell 的重定向语法，不是传给 Agent 的参数。Shell 在启动命令前打开文件，Bubblewrap 通过继承的 FD 4 读取指令。

也可以显式打开文件：

```sh
(
  exec 4< rules.bpf
  bwrap --ro-bind / / --seccomp 4 /absolute/path/to/agent
)
```

括号将文件描述符的生命周期限定在子 Shell 内；退出后不会给原 Shell 留下 FD 4。实际程序调用时，也必须确保启动 Bubblewrap 的接口会传递这个 FD，而不是在执行前关闭它。

上述命令只演示 seccomp 规则传递，仍映射了整个宿主根目录，不构成完整的文件访问隔离策略。示例需替换实际程序路径，并在 Linux 环境运行；本文未执行该示例。

## 4. 完整执行流程

```text
TypeScript 等代码定义系统调用规则
  ↓ 转换为经典 BPF 指令
生成 BPF 二进制文件
  ↓ Shell 或启动程序打开文件
得到并传递 FD 4
  ↓
Bubblewrap 通过 --seccomp 4 读取规则并安装
  ↓
启动 Agent
  ↓
内核对 Agent 及其后续子进程执行过滤
```

文件本身不会对任何进程施加限制；安装过滤器后才生效。过滤器在 `exec` 后保留，也由后续创建的子进程继承，不需要每个后代再次读取文件。

## 5. 使用条件与排查

- 用 `bwrap --help` 检查是否提供 `--seccomp` 选项；内核也需要支持 seccomp 过滤器。
- 确认 FD 有效、可读取，并且读取位置位于过滤指令开始处。
- 文件内容必须符合 Bubblewrap 接受的原始 BPF 指令格式，规则还需适配目标架构和系统调用 ABI。
- 过滤器必须允许目标程序启动和运行所需的系统调用，包括实际使用的执行接口、动态链接器和运行时所需调用。
- 安装失败应按失败处理，不能把未安装过滤器的任务当作已经受到限制。
