# Codex 沙箱

Codex 本地沙箱通过操作系统机制限制执行权限，不默认要求 Docker。本文依据 OpenAI 官方文档整理，沙箱适用范围与启用方式核对日期：2026-11-02；平台实现细节沿用 2026-10-22 的记录。

## 1. 平台实现

| 平台 | 实现 |
| --- | --- |
| macOS | Seatbelt，通过 `sandbox-exec -p` 加载与所选模式对应的策略 |
| Linux、WSL2 | 默认使用 Bubblewrap（`bwrap`）配合 seccomp |
| Windows | 支持 MXC，另有 `elevated`、`unelevated` 后备实现 |

宿主根据权限配置准备策略，再启动受限命令；命令创建的子进程也受执行边界约束。通用机制与 Codex 的配置转换逻辑需要区分。

## 2. 工作目录与配置

在 `config.toml` 中可配置：

```toml
sandbox_mode = "workspace-write"
approval_policy = "on-request"
approvals_reviewer = "user"

[sandbox_workspace_write]
network_access = false
```

`workspace-write` 允许修改真实工作目录及额外授权目录，不是复制仓库后再修改。`read-only` 默认限制写入；`danger-full-access` 移除沙箱限制。

## 3. 与审批的关系

沙箱规定执行边界，审批策略决定如何申请越界权限，审批者可以是用户或自动审核器。获批后由宿主按相应权限执行，不是操作系统沙箱自行展示审批。

自动审核不等于扩大默认边界；`approval_policy = "never"` 也不等于自动获得越界权限。

## 4. 命令与工具的适用范围

官方文档明确说明：

> The sandbox applies to spawned commands, not just to built-in file operations.

沙箱约束宿主启动的命令，`git`、包管理器、测试工具以及它们创建的子进程都会继承相同的执行边界。限制来自运行环境，不是 Bash 自带的能力；使用 `bash`、`zsh`，或者在其中启动 `python`、`npm`，都不能仅凭程序名称判断是否启用了沙箱。

默认权限模式自动启用沙箱，但实际边界取决于本次执行的权限配置。配置为 `danger-full-access` 或获准在沙箱外执行的命令，不受原有沙箱边界约束；不能将“调用了终端工具”直接等同于“必然在沙箱内执行”。

文件操作也受对应文件权限控制，但具体工具由宿主实现。MCP、外部连接器和浏览器操作不能仅因由 Codex 调用，就认为自动纳入了同一个本地命令沙箱；应检查各工具的执行位置、服务权限和独立审批要求。

## 5. 本地验证

官方提供沙箱测试命令，例如：

```sh
codex sandbox macos /bin/pwd
codex sandbox linux /bin/pwd
```

按当前平台选择命令，并确认所用权限配置。本文没有实际执行这些示例。

## 6. 官方文档

- [Sandbox：边界、模式与审批配置](https://learn.chatgpt.com/docs/sandboxing)
- [Agent approvals & security：各平台实现与验证命令](https://learn.chatgpt.com/docs/agent-approvals-security#os-level-sandbox)
