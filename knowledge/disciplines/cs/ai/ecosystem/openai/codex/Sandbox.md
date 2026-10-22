# Codex 沙箱

Codex 本地沙箱通过操作系统机制限制执行权限，不默认要求 Docker。本文依据 OpenAI 官方文档整理，核对日期：2026-10-22。

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

## 4. 本地验证

官方提供沙箱测试命令，例如：

```sh
codex sandbox macos /bin/pwd
codex sandbox linux /bin/pwd
```

按当前平台选择命令，并确认所用权限配置。本文没有实际执行这些示例。

## 5. 官方文档

- [Sandbox：边界、模式与审批配置](https://learn.chatgpt.com/docs/sandboxing)
- [Agent approvals & security：各平台实现与验证命令](https://learn.chatgpt.com/docs/agent-approvals-security#os-level-sandbox)
