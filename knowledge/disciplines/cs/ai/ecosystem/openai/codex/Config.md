# Codex Config

Codex 使用 TOML 配置模型、权限、MCP 等行为。依据 OpenAI 官方文档整理，核对日期：2026-10-22。

## 1. 配置位置

- 用户级：`~/.codex/config.toml`。
- 项目级：项目或子目录的 `.codex/config.toml`，仅受信任项目加载。

## 2. 配置优先级

由高到低：CLI 覆盖 → 项目配置（更接近当前目录的优先）→ 选定的 Profile 文件 → 用户配置 → 云端管理的默认配置 → 系统配置 → 内置默认值。

管理员的 `requirements.toml` 是强制约束，不能把上述优先级理解为 CLI 可以解除组织限制。

## 3. 权限配置入口

| 内容 | 入口 |
| --- | --- |
| 审批触发与审核者 | `approval_policy`、`approvals_reviewer` |
| 沙箱与可写目录 | `sandbox_mode`、`sandbox_workspace_write.writable_roots` |
| 沙箱命令联网 | `sandbox_workspace_write.network_access` |
| MCP 工具审批 | 服务的 `default_tools_approval_mode` 与单工具 `approval_mode` |
| 命令匹配 | `.rules` 文件 |

各入口控制不同层次，工具允许调用不代表解除资源限制。完整字段及适用版本见官方参考。

## 4. 官方文档

- [Config basics：位置、信任与优先级](https://learn.chatgpt.com/docs/config-file/config-basic)
- [Configuration reference：完整配置字段](https://learn.chatgpt.com/docs/config-file/config-reference)
- [Sandbox：执行边界与审批](https://learn.chatgpt.com/docs/sandboxing)
- [MCP：工具审批配置](https://learn.chatgpt.com/docs/extend/mcp)
