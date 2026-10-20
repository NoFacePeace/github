# Claude Code Hooks

Hook 是由 Claude Code 在生命周期事件发生时自动执行的操作，可用于参数校验、格式检查、记录结果和补充上下文。本文根据官方文档整理，核对日期为 2026-10-20。

## 1. 配置位置

| 位置 | 范围 |
|---|---|
| `~/.claude/settings.json` 的 `hooks` | 用户级，适用于多个项目 |
| `<项目>/.claude/settings.json` 的 `hooks` | 项目级，可提交到仓库共享 |
| `<项目>/.claude/settings.local.json` 的 `hooks` | 项目个人设置 |
| 插件内的 `hooks/hooks.json` | 随插件启用和管理 |

Hook 可以独立配置，不要求创建插件。组织还可以通过托管策略配置。脚本的存放位置本身不是注册入口，必须由 Hook 配置引用。

## 2. 触发与处理

```text
宿主加载配置 → 事件发生 → 条件匹配 → 执行处理程序 → 处理返回结果
```

常见事件包括会话开始 `SessionStart`、工具执行前 `PreToolUse`、工具执行后 `PostToolUse` 和停止 `Stop`。工具事件的 `matcher` 可匹配工具名称；其他事件使用各自定义的字段。

处理程序支持命令、HTTP、MCP 工具、提示词和子 Agent 等形式，具体适用事件以官方参考为准。命令型 Hook 从标准输入接收事件 JSON，通过退出状态和标准输出反馈结果。

## 3. 配置示例

以下项目级示例在写入或编辑文件后执行检查。将 `hooks` 合并到 `.claude/settings.json`，保留已有设置：

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "bash",
            "args": ["${CLAUDE_PROJECT_DIR}/.claude/hooks/check-style.sh"]
          }
        ]
      }
    ]
  }
}
```

该示例要求先准备检查脚本。`${CLAUDE_PROJECT_DIR}` 指向会话启动时的项目根目录，避免将当前命令目录误认为脚本目录。这里仅记录配置，不安装或启用任何实际 Hook。

## 4. 与 Skill、工具和插件的关系

- Skill 提供任务流程和资源，由用户或 Agent 使用。
- 工具提供操作接口，通常由模型发起调用。
- Hook 根据宿主事件触发，不依赖模型每次主动选择。
- [插件](Plugins.md) 可将 Hook 与其他扩展一起分发。

Hook 不一定是 Bash 脚本，也不默认把处理程序全文放进上下文。返回数据是否能补充上下文、阻止操作或影响继续执行，取决于事件和处理程序类型。工具执行后的检查不能被假设为能撤销已经完成的操作。

## 5. 官方文档

- [Hooks：事件、配置、输入输出与退出状态](https://code.claude.com/docs/en/hooks)
- [插件清单中的 Hook 配置](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code 扩展机制](https://code.claude.com/docs/en/features-overview)
