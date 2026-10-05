# Mods

## 1. 基础概念

Mods 是 Claude Code 的程序扩展机制。开发者可以用 JavaScript 或 TypeScript 编写扩展，并作为插件加载，在 Claude Code 运行过程中的指定事件上执行逻辑，调整提示词、工具调用、权限、子 Agent 和界面行为。

模型负责推理和决策，Claude Code 运行系统负责触发事件与执行 Mod 代码。因此，Mods 属于 Claude Code 产品的扩展能力，不是安装后就能在任意模型 API 中使用的能力。

### 1.1 与 Skills 和 MCP 的区别

| 机制 | 主要作用 | 示例 |
| --- | --- | --- |
| Skills | 向模型提供任务知识和操作说明 | 指导模型按团队规范审查代码 |
| MCP | 向 Agent 提供外部工具和资源 | 查询数据库、读取业务文档 |
| Mods | 用代码介入 Claude Code 的运行流程 | 检查工具参数、修改提示词、控制子 Agent 启动 |

Mods 的事件主要发生在 Claude Code 内部，与外部系统推送消息的 MCP Events 是不同机制。

## 2. 使用方式

### 2.1 版本与加载

官方入门资料要求 Claude Code 2.1.287 或更高版本，Mods 默认启用。安装并启用包含 Mod 的插件后，Claude Code 会加载对应扩展。

是否立刻产生效果取决于扩展的设计：监听事件的扩展会在事件发生时执行；提供命令或配置项的扩展，需要用户调用命令或设置配置。

### 2.2 创建与本地使用

可以启动 Claude Code，让它帮助创建 Mod，例如：“创建一个 Mod，在执行 Bash 工具前检查命令。”按提示加载后可以在当前会话中测试；希望后续会话继续使用，需要保存并安装插件。

已有本地插件目录时，可以通过启动参数加载：

```bash
claude --plugin-dir ./my-mod
```

这里的 `my-mod` 应是包含 Mod 的有效插件目录。

### 2.3 从插件市场安装

在 Claude Code 中运行以下命令，将占位符替换为实际市场和插件名称：

```text
/plugin marketplace add <org>/<repo>
/plugin install <plugin>@<marketplace>
/reload-plugins
```

## 3. 支持的事件

以下列表依据官方 v2.1.289 事件参考。不同版本可能存在差异，具体事件名称、参数和处理函数签名应以安装版本生成的 TypeScript 类型声明为准。

### 3.1 工具

| 事件 | 用途 |
| --- | --- |
| `tool.call` | 工具即将执行时触发，可观察、修改参数、拒绝调用或直接返回结果 |
| `tool.check` | 工具权限判断，可返回允许、询问或拒绝；发生在 `tool.call` 和 `PreToolUse` 之后 |
| `tool.describe` | 介入工具描述 |

### 3.2 提示词与上下文

相关事件包括 `prompt.submit`、`prompt.fill`、`prompt.suggest`、`prompt.edit`、`prompt.compose`、`prompt.section`、`prompt.context`、`prompt.attachment`、`skill.prompt` 和 `attribution.text`。

它们覆盖输入提交、编辑、建议、提示词组装、上下文、附件、Skill 提示词和归属文本等环节。具体可修改的字段由各事件的类型定义决定。

### 3.3 命令与配置

相关事件包括 `command.run`、`command.describe`、`config.set` 和 `config.describe`，用于介入命令执行、命令描述、配置设置和配置描述。

### 3.4 对话轮次

| 事件 | 用途 |
| --- | --- |
| `turn.start` | 一轮处理开始 |
| `turn.step` | 每次请求模型之前，可调整模型和推理 effort |
| `turn.complete` | 一轮处理完成 |

一轮处理可能包含多次模型请求与工具调用，所以 `turn.step` 可以在同一轮中触发多次。

### 3.5 会话

相关事件包括 `session.start`、`session.end`、`session.compact`、`session.receive`、`session.send`、`session.append`、`session.attach`、`session.detach` 和 `session.measure`，覆盖会话生命周期、上下文压缩、消息处理、附加内容及测量等环节。

### 3.6 子 Agent

| 事件 | 用途 |
| --- | --- |
| `agent.offer` | 介入可供使用的 Agent |
| `agent.spawn` | 子 Agent 或团队成员启动前触发，可选择模型或拒绝启动 |

Mod 在这些节点执行程序逻辑，后续 Agent 的实际启动和运行由 Claude Code 负责。

### 3.7 界面

相关事件包括 `ui.render`、`ui.resolve`、`ui.press`、`ui.input`、`ui.select`、`ui.focus`、`ui.scroll`、`ui.close`、`ui.message` 和 `ui.fault`，覆盖界面渲染、交互、消息和错误等环节。

### 3.8 插件与遥测

- 插件与引擎：`plugin.register`、`engine.create`。
- 遥测：`telemetry.log`、`telemetry.mark`。用户安装的 Mod 监听遥测时必须使用 `{ to: "collector" }` 过滤条件，通配符 `*` 不会匹配这些事件。

### 3.9 传统 Hooks 与 API 调用

Mods 也支持 `classic.<Event>` 形式的传统 Hook 事件，例如 `classic.Stop` 和 `classic.PostToolUse`。

Mods API 方法调用也可作为事件，例如 `fs.read`、`model.complete`、`ui.open` 和 `process.spawn`，可用于拦截在其后加载的 Mod 发出的调用。不要将其理解为一定能拦截 Claude Code 内部的所有同类操作。

## 4. 事件处理示例

下面的示例监听 Bash 工具调用，然后放行原始事件：

```javascript
on("tool.call", { tool: "Bash" }, async ($, e, next) => {
  // 将原始事件传给后续处理流程。
  return next(e);
});
```

这是事件处理函数片段，需要放入有效的 Mod 插件中使用。

一般处理函数可以观察事件、复制并修改事件后调用 `next`，或按该事件约定直接回答、拒绝，不再继续调用 `next`。因此，Mods 不只是接收通知，还能影响后续行为。`turn.step` 和 `process.spawn` 使用异步生成器处理函数，不能直接套用上述普通异步函数签名。

## 5. 运行环境与边界

Mod 模块在受限的 JavaScript 环境中运行，不直接提供 Node.js 或 DOM 全局能力；文件、进程、模型等能力通过 Mods 提供的 `$` API 使用。

模块隔离不等于插件可以被当作不可信代码安全运行。扩展能够影响 Claude Code 的行为并通过提供的 API 使用外部能力，安装前应了解其代码和功能。

## 6. 参考资料

- [Claude Code Mods 入门](https://claude.dev/blog/getting-started-with-claude-code-mods/)
- [创建 Mods](https://code.claude.com/docs/en/plugins/mods/create)
- [Mods API 与事件参考](https://code.claude.com/docs/en/plugins/mods/reference)
- [Claude Code Mods README](https://github.com/anthropics/claude-code/blob/main/mods/README.md)

整理日期：2026 年 10 月 5 日。版本要求与事件支持范围应以当前官方文档及本地类型声明为准。
