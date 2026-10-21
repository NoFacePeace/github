# Agent Hooks

钩子（Hook）是 Agent 宿主在特定生命周期事件发生时，按配置触发操作的扩展机制。例如，在工具执行前校验调用，在工具执行后运行检查，或在会话开始时补充上下文。

## 1. 执行流程

```text
宿主加载 Hook 配置
  → 运行过程中产生事件
  → 匹配事件类型和条件
  → 执行 Hook 处理程序
  → 按事件约定处理返回结果
  → 继续执行、阻止操作或补充上下文
```

Hook 由宿主根据事件触发，不需要模型每次主动选择调用。具体支持哪些事件、处理程序和返回结果，由宿主定义。

## 2. 常见用途

| 触发时机 | 用途示例 |
|---|---|
| 会话开始 | 准备环境、补充项目说明 |
| 用户提交消息 | 校验输入、补充任务上下文 |
| 工具执行前 | 检查参数、限制操作 |
| 工具执行后 | 格式化文件、记录结果、运行检查 |
| Agent 停止 | 检查完成条件、发送通知 |

不同事件允许的行为不同。例如，工具执行后运行检查，不代表能够撤销已经发生的操作。Hook 输出也不一定进入模型上下文，只有宿主明确支持的输出字段或文本才会被使用。

## 3. 与其他机制的关系

| 机制 | 主要职责 | 触发方式 |
|---|---|---|
| Hook | 在生命周期节点执行预设操作 | 宿主按事件和条件触发 |
| Skill | 提供任务流程、知识和资源 | 用户调用或 Agent 根据任务使用 |
| 工具 | 提供具体的可调用操作 | 模型提出调用请求，由宿主执行 |
| 插件 | 打包和分发扩展能力 | 宿主安装、启用和加载 |

Hook 可以独立配置，也可以随插件分发。脚本只是处理程序的一种实现；仅放置脚本文件不会自动注册 Hook，还需要配置事件、匹配条件和执行入口。

相关内容：[工具](../tools/Tools.md)、[插件](../plugins/Plugins.md)、[事件](../../execution/events/Events.md)。

## 4. Codex 与 Claude Code

以下位置与行为依据官方文档整理，核对日期为 2026-10-20；具体支持范围可能随版本变化。

| 配置范围 | Codex | Claude Code |
|---|---|---|
| 用户级 | `~/.codex/hooks.json`，或 `~/.codex/config.toml` 中的 `[hooks]` | `~/.claude/settings.json` 中的 `hooks` |
| 项目级 | `<repo>/.codex/hooks.json`，或 `<repo>/.codex/config.toml` 中的 `[hooks]` | `<repo>/.claude/settings.json` 中的 `hooks` |
| 项目级个人设置 | 此表不指定额外入口 | `<repo>/.claude/settings.local.json` 中的 `hooks` |
| 插件附带 | 默认发现插件内的 `hooks/hooks.json`，也可通过插件清单指定 | 插件内的 `hooks/hooks.json`，也可通过插件清单声明 |

两者都支持命令型 Hook，但不能由此推断所有事件、字段和返回值都兼容。迁移配置时应核对对应平台文档。

Codex 对普通用户、项目和插件提供的非托管 Hook 要求审阅并信任具体定义。CLI 中可通过 `/hooks` 查看、审阅和信任；新建或修改的定义在获得信任前会被跳过。项目级 Hook 还要求项目配置层受信任。管理员托管的 Hook 可由策略信任。

信任表示用户认可这条自动执行规则，不代表宿主已经验证脚本安全，也不等于授予所有操作无限权限。

## 5. 参考资料

- [Codex Hooks](https://learn.chatgpt.com/docs/hooks)
- [Codex 插件打包与 Hook 配置](https://developers.openai.com/plugins/build/plugins)
- [Claude Code Hooks](https://code.claude.com/docs/en/hooks)
- [Claude Code 插件清单参考](https://code.claude.com/docs/en/plugins-reference)
