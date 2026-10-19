# Arize Phoenix

Phoenix 是 Arize AI 开发的 AI 应用可观测性与评估平台，用于查看 LLM 和 Agent 的执行过程、定位问题，以及比较不同模型、提示词和流程的效果。

## 1. 核心能力

| 能力 | 用途 |
|---|---|
| 执行追踪 | 查看模型、检索和工具调用的输入、输出、耗时与错误 |
| 数据集与实验 | 保存案例，比较不同版本在同一批任务上的表现 |
| 评估与标注 | 通过人工标注、代码检查或模型评分检验输出质量 |

例如，Agent 回答订单状态有误时，可以沿执行轨迹检查工具参数、查询结果和模型回复，定位错误出现在哪一步。

## 2. 与 Agent 知识的关系

- [可观测性](../../agent/observability/Observability.md)：Phoenix 收集和展示执行轨迹，帮助理解实际运行过程及失败原因。
- [评估](../../agent/evaluation/Evaluation.md)：Phoenix 提供数据集、实验和评分能力，帮助验证任务效果及版本改进。

追踪不会自动覆盖所有业务步骤，应用需要接入对应的埋点组件；评估也需要按任务定义标准，不能仅凭调用成功判断目标已经完成。

## 3. 本地使用流程

### 3.1 启动服务

安装 uv 后，可以在终端启动 Phoenix：

```bash
uvx arize-phoenix serve
```

保持进程运行，在浏览器打开 `http://localhost:6006`。也可以在 Python 虚拟环境中安装 `arize-phoenix`，再运行 `phoenix serve`。

### 3.2 接入应用

在 Agent 项目目录中运行官方接入向导：

```bash
npx @arizeai/phoenix-cli setup
```

向导协助识别框架和模型服务，配置对应的 OpenInference 埋点组件及轨迹上报。启动 Phoenix 本身不会自动采集其他应用的对话。

也可以手动接入。例如，Python 应用安装 `arize-phoenix-otel` 及所用框架对应的 `openinference-instrumentation-*` 包后，在程序启动时配置：

```python
from phoenix.otel import register

register(
    project_name="my-agent",
    endpoint="http://localhost:6006/v1/traces",
    protocol="http/protobuf",
    auto_instrument=True,
)
```

自动追踪范围由已安装且受支持的埋点组件决定。自定义业务步骤可以补充手动埋点。

### 3.3 查看与评估

1. 正常运行 Agent，向其提交一个任务。
2. 在 Phoenix 对应项目中查看轨迹，检查模型及工具调用的输入、输出和异常。
3. 保存有代表性的案例并整理为数据集，配置评分标准。
4. 修改模型、提示词或流程后重新运行实验，对比结果并检查回归。

## 4. 部署与许可

Phoenix 支持本地运行和服务器自托管，源码公开，采用 Elastic License 2.0（ELv2）。官方使用“open-source”描述项目，但 ELv2 是带有限制的源码可用许可证，不等同于 MIT 或 Apache 2.0；使用和对外提供服务前应查看许可条款。

运行 Phoenix 服务不要求同时在本机加载大模型。应用可以调用云端模型，也可以接入本地模型；使用模型评分时，所选模型服务可能产生额外费用。

## 5. 参考资料

- [官方介绍](https://arize.com/phoenix/)
- [官方仓库与启动说明](https://github.com/arize-ai/phoenix)
- [官方文档](https://arize.com/docs/phoenix/)
- [追踪配置](https://arize.com/docs/phoenix/tracing/how-to-tracing/setup-tracing/custom-spans)
- [许可说明](https://github.com/Arize-ai/phoenix/blob/main/docs/phoenix/self-hosting/license.mdx)
