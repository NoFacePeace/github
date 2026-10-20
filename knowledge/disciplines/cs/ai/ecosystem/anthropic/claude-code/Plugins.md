# Claude Code 插件

插件将技能、命令、子 Agent、Hook 和外部服务连接等扩展打包，便于复用与分发。本文根据官方文档整理，核对日期为 2026-10-20。

## 1. 目录与组件

以下为可选组件的组合示例，不要求每个插件包含所有文件：

```text
my-plugin/
├── .claude-plugin/
│   └── plugin.json
├── skills/
│   └── review/
│       ├── SKILL.md
│       ├── scripts/
│       └── references/
├── commands/
│   └── check.md
├── agents/
│   └── reviewer.md
├── hooks/
│   └── hooks.json
├── .mcp.json
├── .lsp.json
└── output-styles/
```

| 入口 | 作用 |
|---|---|
| `.claude-plugin/plugin.json` | 插件名称、版本、组件路径等清单信息 |
| `skills/` | 技能说明及配套资源 |
| `commands/` | Markdown 命令定义 |
| `agents/` | 子 Agent 定义 |
| `hooks/hooks.json` | 生命周期事件处理配置，详见 [Hooks](Hooks.md) |
| `.mcp.json` | MCP 工具服务连接配置 |
| `.lsp.json` | 语言服务器配置 |
| `output-styles/` | 输出风格定义 |

组件放在插件根目录下，不放入 `.claude-plugin/`。`scripts/` 和 `references/` 通常服务于 Skill，也可由其他组件引用；不会自动成为工具或全部进入模型上下文。

## 2. 清单与发现

最小清单示例：

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Reusable development workflows"
}
```

当前官方文档允许没有清单时按标准目录发现组件；清单用于补充元数据、自定义路径、内联定义和用户配置。指定组件路径时，应遵守官方的相对路径与目录边界规则。

自定义路径与默认位置的组合规则因组件而异：例如 `skills` 添加额外扫描位置，`commands` 和 `agents` 替换默认扫描入口，`hooks` 与默认 Hook 文件共同加载。不能统一理解为覆盖或追加。

## 3. 验证与执行

创建插件后，可通过官方校验命令检查清单与引用路径：

```sh
claude plugin validate ./my-plugin
```

宿主发现各组件后，按组件规则提供技能、命令、子 Agent 或工具入口。模型需要使用脚本时，通过已有执行工具运行；MCP 工具通过对应服务执行；Hook 则由宿主事件触发。

## 4. 官方文档

- [插件清单、路径规则与标准布局](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code 扩展机制](https://code.claude.com/docs/en/features-overview)
- [Hooks 配置与执行规则](https://code.claude.com/docs/en/hooks)
