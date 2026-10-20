# Codex Hooks

Hook 由 Codex 宿主在生命周期事件发生时触发，可用于校验操作、记录结果或补充上下文。无需先让模型决定调用。通用概念见 [Agent Hooks](../../../agent/hooks/Hooks.md)。本文依据官方文档整理，核对日期为 2026-10-20。

## 1. 配置入口

| 范围 | 入口 |
|---|---|
| 用户级 | `~/.codex/hooks.json`，或 `~/.codex/config.toml` 中的 `[hooks]` |
| 项目级 | `<repo>/.codex/hooks.json`，或 `<repo>/.codex/config.toml` 中的 `[hooks]` |
| 插件级 | 默认发现插件根目录下的 `hooks/hooks.json`，或由清单指定 |

独立 Hook 不需要插件。多个来源的匹配 Hook 会共同加载，不是高优先级来源整体覆盖低优先级来源。同一配置层同时使用 JSON 与内联 TOML 时会合并并产生警告，通常每层选择一种表示。

## 2. 事件与处理程序

配置分为三层：事件、匹配组、处理程序。常见事件包括 `SessionStart`、`UserPromptSubmit`、`PreToolUse`、`PostToolUse` 和 `Stop`；完整列表及每个事件的字段以官方文档为准。

`matcher` 的匹配对象随事件不同。例如，`SessionStart` 匹配启动来源，工具事件匹配工具名称。当前文档支持 `command` 和 `mcp_tool` 处理程序；`prompt` 和 `agent` 处理程序虽然可解析，但会被跳过。

## 3. 最小配置示例

以下用户级配置在新会话开始时运行脚本，保存为 `~/.codex/hooks.json`：

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "^startup$",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.codex/hooks/session-start.sh"
          }
        ]
      }
    ]
  }
}
```

先创建被引用的脚本，并完成信任审阅，再通过新会话验证。这里展示的是文档配置示例，不会修改当前机器的 Hook 设置。

命令型处理程序通过标准输入接收事件 JSON，以退出状态及事件支持的输出格式反馈结果。不要把命令工作目录当作脚本所在目录；项目脚本应使用明确路径或从仓库根目录解析。

## 4. 审阅与信任

普通用户、项目和插件提供的非托管 Hook 需要先审阅并信任：

1. 在 Codex CLI 中输入 `/hooks`。
2. 检查来源、事件、匹配条件和执行命令。
3. 信任需要启用的定义。

信任绑定当前 Hook 定义的哈希；新建或修改定义后，在重新信任前会跳过。项目级 Hook 还要求项目 `.codex/` 配置层受信任。系统、MDM 或云策略管理的 Hook 可由组织策略信任。

信任自动执行规则不代表脚本经过安全认证，也不表示运行权限不受限制。

## 5. 插件 Hook 与上下文

便携式插件通过根 `plugin.json` 的 `extensions.com.openai.hooks` 指定 Hook；旧格式可在 `.codex-plugin/plugin.json` 中声明。未显式指定时，宿主默认寻找 `hooks/hooks.json`；显式配置替换该默认发现入口。

插件命令可使用 `${PLUGIN_ROOT}` 定位安装目录，使用 `${PLUGIN_DATA}` 定位插件可写数据目录。安装或启用插件与信任 Hook 是不同步骤。

Hook 不默认把脚本正文放进模型上下文。事件允许时，处理程序可以返回额外上下文；例如 `SessionStart` 的纯文本输出或 `hookSpecificOutput.additionalContext` 会加入 developer 上下文。工具执行后发生的 Hook 也不能被假设为能撤销已完成操作。

## 6. 官方文档

- [Hooks：配置来源、事件、输入输出及信任](https://learn.chatgpt.com/docs/hooks)
- [插件打包：附带生命周期 Hook](https://developers.openai.com/plugins/build/plugins)
- [配置参考](https://learn.chatgpt.com/docs/config-file/config-reference)
- [高级配置与项目配置层](https://learn.chatgpt.com/docs/config-file/config-advanced)
