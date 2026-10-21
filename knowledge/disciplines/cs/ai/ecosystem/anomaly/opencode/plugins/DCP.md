# DCP

DCP（Dynamic Context Pruning）是 OpenCode 的社区上下文管理插件，通过模型可调用的压缩工具与自动清理，减少后续模型请求携带的内容。它不属于 Anomaly 官方产品。

## 1. 模型主动压缩

`compress` 工具允许模型选择处理范围，并在参数中提交摘要。插件校验选择、处理保护内容并记录替换状态。支持连续范围压缩，以及实验性的单消息压缩。

## 2. 请求前替换

插件通过消息转换钩子在请求发送前应用规则，跳过被压缩的原消息，在锚点插入摘要；原始会话历史保留。

```text
原始历史：A → B → C → D → E
替换规则：B～D 由摘要 S 替代
模型输入：A → S → E
```

压缩修改的是下一次模型输入，不是已经进行中的请求。

## 3. 自动清理与保护

项目还提供同工具同参数的旧结果去重，以及错误调用输入清理。可以配置受保护工具、文件和用户消息。具体清理时机与保护范围应按版本核对，不能将同参数直接等同于结果相同。

## 4. 项目状态

截至 2026-10-21 核对，README 表示 DCP 开发放缓，新增上下文管理工作主要转向 Sleev。DCP 仍可作为 OpenCode 插件使用；此前资料中的 `discard`、`extract` 等工具名称不宜直接套用到当前 `compress` 接口。

## 5. 参考资料

- [DCP 项目与 README](https://github.com/Tarquinen/opencode-dynamic-context-pruning)
- [compress 范围模式](https://github.com/Tarquinen/opencode-dynamic-context-pruning/blob/master/lib/compress/range.ts)
- [消息转换钩子](https://github.com/Tarquinen/opencode-dynamic-context-pruning/blob/master/lib/hooks.ts)
- [请求消息替换](https://github.com/Tarquinen/opencode-dynamic-context-pruning/blob/master/lib/messages/prune.ts)
