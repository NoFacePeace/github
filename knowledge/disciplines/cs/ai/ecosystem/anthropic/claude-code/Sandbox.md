# Claude Code 沙箱

Claude Code 提供由操作系统强制执行的本地 Shell 沙箱，可限制命令访问的文件路径和网络域名。本文依据官方文档整理，核对日期：2026-11-02。

## 1. 启用方式

本地 Shell 沙箱默认关闭。在 Claude Code 会话中运行 `/sandbox` 开启，也可以在 `~/.claude/settings.json` 中配置：

```json
{
  "sandbox": {
    "enabled": true
  }
}
```

macOS 使用系统内置的 Seatbelt；Linux、WSL2 依赖 Bubblewrap 和 socat。原生 Windows 当前不支持此沙箱，可在 WSL2 中使用。

## 2. 命令与工具的适用范围

官方文档明确说明：

> While the sandbox is on, the shell commands Claude runs start inside its boundary, and so do the processes they start.

启用后，Claude 执行的 Shell 命令及其子进程在沙箱边界内运行。沙箱覆盖 Bash、PowerShell、Monitor 命令及其启动的进程，因此 `git`、`npm`、`python`、构建和测试程序也会继承限制。这里的 PowerShell 指支持沙箱的平台上的命令，不表示原生 Windows 支持此沙箱。

以下能力不在这个 Shell 沙箱内：

| 能力 | 控制方式或执行范围 |
| --- | --- |
| `Read`、`Edit`、`Write` 等内置文件工具 | 遵循独立权限规则，不能用 Shell 沙箱配置推断其访问范围 |
| `WebFetch`、`WebSearch` | 不受 Shell 沙箱的网络域名白名单直接约束 |
| 本地 MCP 服务、Hooks、LSP 等辅助进程 | 在 Shell 沙箱外运行，使用宿主进程可用的权限 |

如果需要将整个 Claude Code 进程及这些工具放在统一隔离边界内，应另行使用容器、虚拟机或外层沙箱。

## 3. 文件与网络边界

- 默认可写当前工作目录、用户临时目录及额外添加的目录，部分受保护路径仍禁止写入。
- 默认可读机器上的大多数文件；启用沙箱本身不意味着禁止读取 SSH、云服务凭据等敏感文件，需要额外配置读取限制。
- 网络连接经过本地代理，按允许的域名控制访问；网络授权和审批行为取决于权限模式。

## 4. 与审批及沙箱外执行的关系

沙箱控制命令在操作系统层面能访问什么，权限模式决定是否需要审批。自动批准沙箱内命令不会移除文件和网络限制。

匹配 `excludedCommands` 的命令可在沙箱外执行；命令受阻后，也可能按权限规则获准在沙箱外重试。因此，“调用 Bash 工具”不等于“本次命令一定在沙箱内”。

可以通过以下配置启用沙箱、禁止沙箱外重试，并在沙箱不可用时停止启动：

```json
{
  "sandbox": {
    "enabled": true,
    "allowUnsandboxedCommands": false,
    "failIfUnavailable": true
  }
}
```

这不意味着所有工具都进入了沙箱，仍需检查 `excludedCommands` 等配置。用户自己在 `!` Shell 模式输入的命令通常也在沙箱外执行，不适合用来验证 Claude 发起的命令是否受限。

## 5. 官方文档

- [Sandbox：配置、平台支持与边界](https://code.claude.com/docs/en/sandboxing)
- [What the sandbox restricts：命令及子进程的适用范围](https://code.claude.com/docs/en/sandboxing#what-the-sandbox-restricts)
- [What runs outside the sandbox：文件工具、MCP 与 Hooks 等例外](https://code.claude.com/docs/en/sandboxing#what-runs-outside-the-sandbox)
- [Strict sandbox mode：禁止沙箱外重试](https://code.claude.com/docs/en/sandboxing#turn-off-the-retry-with-strict-sandbox-mode)
