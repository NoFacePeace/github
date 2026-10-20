# Codex 插件

插件用于打包和分发 Skill、MCP 服务连接及相关资源。ChatGPT 与 Codex 共享插件目录，但具体能力取决于运行环境。本文依据官方文档整理，核对日期为 2026-10-20。

## 1. 插件结构

官方当前推荐新包使用根目录的便携式清单；旧包仍可能采用 Codex 兼容清单。

```text
my-plugin/
├── plugin.json                  # 推荐的便携式清单
├── skills/
│   └── my-skill/
│       ├── SKILL.md
│       ├── scripts/             # 可选的任务脚本
│       └── references/          # 可选的参考资料
├── mcp.json                     # 可选的便携式 MCP 配置
├── .codex-plugin/
│   └── plugin.json              # 可选的 Codex 兼容清单
├── hooks/
│   └── hooks.json               # 可选的生命周期 Hook
└── assets/                      # 图标、模板等资源
```

便携格式使用固定的 `skills/`、`mcp.json` 入口；OpenAI 专属设置放在根清单的 `extensions.com.openai` 中。文档说明：缺少该扩展对象时，可以使用 `.codex-plugin/plugin.json` 作为兼容回退。两份清单不应被理解为无条件叠加。

兼容清单中常见字段包括 `skills`、`mcpServers`、`apps` 和 `interface`，分别用于声明技能目录、MCP 配置、已注册服务的应用映射和界面信息。旧包常见的 `.mcp.json`、`.app.json` 位于插件根目录，不放在 `.codex-plugin/` 内。

## 2. 安装与配置位置

| 信息 | 位置或作用 |
|---|---|
| 个人本地插件目录清单 | `~/.agents/plugins/marketplace.json` |
| 仓库本地插件目录清单 | `<repo>/.agents/plugins/marketplace.json` |
| 安装副本 | `~/.codex/plugins/cache/<来源>/<插件名>/<版本>/` |
| 用户启用设置 | `~/.codex/config.toml` |
| 项目启用设置 | `<repo>/.codex/config.toml` |

本地 marketplace 插件安装副本的版本目录为 `local`。宿主加载安装后的副本，因此编辑源目录不应被假设为立即更新已安装插件。

可通过 CLI 添加和查看 marketplace：

```sh
codex plugin marketplace add owner/repo
codex plugin marketplace list
```

添加来源不等于安装其中所有插件。官方指南使用桌面应用的插件目录完成本地插件安装和测试，安装后在新聊天中验证。

## 3. 加载与模型上下文

```text
定位已安装插件
  → 解析清单和启用设置
  → 发现技能、连接工具服务
  → 提供技能元数据与工具定义
  → 按任务读取 Skill、调用工具
```

这是加载机制的概念归纳，不表示已核实宿主内部的精确执行顺序。插件代码不会整体装进模型；模型主要看到技能说明、工具接口与执行结果。

Skill 引用的脚本可以通过已有终端工具执行，不会因文件存在而自动注册成独立 Tool。插件只包含说明和资源时，也不一定需要安装脚本。

## 4. Hook 与兼容性

插件可附带 [Hook](Hooks.md)，但安装或启用插件不会自动信任 Hook。运行脚本还要求脚本及依赖存在于执行环境；网页安装不会自动将脚本部署到本机或其他执行环境。

从 Claude Code 迁移时，不能假定 `commands/`、`agents/` 会原样加载。官方迁移指南要求将可复用命令和 Agent 流程转成 Skills；公开目录提交与桌面手动安装对 Hook 的限制也不同。

## 5. 官方文档与示例

- [插件架构](https://developers.openai.com/plugins/concepts/plugins)
- [打包、marketplace 与安装路径](https://developers.openai.com/plugins/build/plugins)
- [编写 Skills](https://developers.openai.com/plugins/build/skills)
- [迁移 Claude Code 插件](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [官方插件示例仓库](https://github.com/openai/plugins)
- [GitHub 插件清单示例](https://github.com/openai/plugins/blob/main/plugins/github/.codex-plugin/plugin.json)
