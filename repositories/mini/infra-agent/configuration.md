# Configuration

| 分类 | 配置项 | 类型 | 默认值或当前值 | 作用及约束 |
| --- | --- | --- | --- | --- |
| 服务 | HTTP 监听地址 | `string` | `:8080` | 服务监听地址，当前在 `main.go` 中指定 |
| 服务 | HTTP 关闭超时 | `time.Duration` | `5s` | 等待存量 HTTP 请求完成的时间 |
| 遥测 | 服务名 | `string` | `infra-agent` | 标识日志、指标与追踪所属服务 |
| 遥测 | 日志导出地址 | `string` | `http://localhost:3100/otlp/v1/logs` | OTLP HTTP 日志接收端点 |
| 遥测 | 关闭超时 | `time.Duration` | `5s` | 刷出剩余遥测数据的时间 |
| 工作区 | `DataDirectory` | `string` | 当前工作目录下 `data` 的绝对路径 | 字段本身无默认值，由入口赋值；内置工作目录为 `<DataDirectory>/workspaces/<WorkspaceID>`，默认工作区 ID 为 `default` |
| 任务 | `DefaultModel` | `string` | 空字符串 | 默认模型标识，使用 `provider/model` 格式；尚未配置默认模型 |
| 任务 | `AutoTitle` | `bool` | `false` | 是否异步生成任务标题；模型生成方法尚待实现 |
| 任务 | `MaxTitleRunes` | `int` | `0` | 标题最大 Unicode 码点数，非正数表示不限制 |
| 任务 | `TitleTimeout` | `time.Duration` | `30s` | 标题生成及更新的超时时间；非正数使用默认值 |
| 执行 | `LeaseTTL` | `time.Duration` | `30s` | 执行租约有效期；非正数使用默认值，续约间隔为有效期的三分之一 |
| 执行 | `CleanupTimeout` | `time.Duration` | `5s` | 执行收尾、释放租约和取消请求存储操作的超时；非正数使用默认值 |
| 执行（规划） | `MaxSteps` | `int` | `200`（建议值） | 主 Run 最大模型调用轮数，由外层传入 `agentloop.Input.MaxSteps`；一次回复包含多个工具调用仍算一步；最后一步禁用工具调用，要求返回最终回复；默认配置尚未接入 |
| 委派（规划） | `ChildMaxSteps` | `int` | `30`（建议值） | 委派子 Run 最大模型调用轮数，执行子 Run 时传入 `agentloop.Input.MaxSteps`；委派及配置尚未实现 |
| 上下文预算（规划） | `TokenBudgetPercent` | `int` | `85`（建议值） | 按模型上下文窗口的百分比计算 `TokenBudget`；配置字段和计算逻辑尚未实现 |
| Token 估算（规划） | `ASCIIBytesPerToken` | `float64` | `4`（建议值） | 每个 Token 对应的 ASCII 字节数，估算为 `ceil(ASCII 字节数 / ASCIIBytesPerToken)`；必须为有限正数，配置及估算逻辑尚未接入 |
| Token 估算（规划） | `NonASCIITokensPerRune` | `float64` | `1.5`（建议值） | 每个非 ASCII Unicode 码点对应的 Token 数，估算为 `ceil(非 ASCII 码点数 × NonASCIITokensPerRune)`；必须为有限正数，与 ASCII 估算相加得到文本估算值，配置及估算逻辑尚未接入 |
| 上下文裁剪 | `TrimTriggerPercent` | `int` | `80` | 预测输入 Token 超过 `Input.TokenBudget × 80%` 时触发裁剪；有效范围为 1～100，范围外使用默认值 |
| 上下文裁剪 | `TrimRetentionPercent` | `int` | `50` | 目标保留去重、截短后消息字节数的 50%，不保证达到目标；有效范围为 1～100，范围外使用默认值 |
| 单次执行参数 | `TokenBudget` | `int` | 无 | 属于 `agentloop.Input`；由外层传入，规划按“模型上下文窗口 × TokenBudgetPercent / 100”计算，作为裁剪阈值的计算基准 |
| 单次执行参数 | `MaxOutputTokens` | `int` | 无 | 属于 `agentloop.Input`；每次模型请求的输出 Token 上限 |
| 历史摘要（规划） | `SummaryTriggerPercent` | `int` | `80`（建议值） | 创建新 Run 时，已有摘要与尚未被摘要覆盖的已结束历史根 Run 消息的估算 Token 总数超过 `TokenBudget × SummaryTriggerPercent / 100`，触发摘要；配置及逻辑尚未实现 |
| 历史摘要（规划） | `SummaryKeepTokens` | `int` | `16000`（建议值） | 保留最近约 16,000 Token 原文，较早内容生成摘要；从末尾按完整消息块累加正文、推理内容及工具调用名称、参数各自的字节数除以 4 的估算值，不拆开工具调用与结果，最新块本身超限时仍保留；此值不包含生成的摘要，摘要输出上限尚未定义，配置及逻辑尚未实现 |
| 安全策略（待实现） | 工具、联网和路径策略 | 未定义 | 未定义 | 已有安全评估接口，配置及具体评估实现尚未接入 |
