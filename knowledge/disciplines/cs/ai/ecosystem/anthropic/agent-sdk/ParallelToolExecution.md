# 并行工具执行

Claude Agent SDK 支持在单个 Agent 的同一轮中并行执行符合条件的工具调用，以减少等待时间。本文依据官方文档整理，核对日期为 2026-11-01。

## 1. 调度规则

当模型在同一轮请求多个工具调用时，运行时根据工具类型决定并行或顺序执行。官方文档给出的规则如下。

| 工具类型 | 执行方式 |
|---|---|
| `Read`、`Glob`、`Grep` 等只读工具 | 可以并行执行 |
| 标注为只读的 MCP 工具 | 可以并行执行 |
| `Edit`、`Write`、`Bash` 等工具 | 顺序执行，避免状态修改冲突 |
| 自定义工具 | 默认顺序执行；通过 `annotations.readOnlyHint` 声明只读后，可以并行执行 |

Python 和 TypeScript SDK 都使用 MCP 的 `readOnlyHint` 字段名。该标记用于描述工具行为，应仅用于实际只读的工具；它不会自动将一个有写入行为的函数变成只读操作，也不替代权限控制。

## 2. 示例与依赖关系

模型在同一轮要求读取 `auth.ts`、`auth.test.ts` 和 `package.json`，这些只读调用可以并行执行，减少等待时间。

```text
模型请求 ─┬→ Read(auth.ts) ──────────┐
          ├→ Read(auth.test.ts) ─────┼→ 工具结果交回模型
          └→ Read(package.json) ────┘
```

“修改代码后运行测试”存在先后依赖，测试应观察修改完成后的代码，因此需要按顺序完成。并行能力本身不会消除任务的数据或状态依赖。

这里的并行指单个 Agent 的一轮响应中多个工具调用的并行，不等于启动多个 Agent。子 Agent 是另一种任务分工机制。

## 3. 参考资料

- [Parallel tool execution](https://code.claude.com/docs/en/agent-sdk/agent-loop#parallel-tool-execution)
- [Python 工具定义参考](https://code.claude.com/docs/en/agent-sdk/python#tool)
- [TypeScript 工具定义参考](https://code.claude.com/docs/en/agent-sdk/typescript#tool)
