# Agent 评估

Agent 评估（Evaluation）是通过明确的任务、场景和标准，判断 Agent 能否正确完成目标，并衡量修改是否带来改进或退步的过程。

## 1. 内容范围

- 测试场景与数据集：定义任务、输入、环境、预期结果和完成条件。
- 评估指标：任务成功率、回答质量、工具使用正确性、耗时和成本等。
- 评分方法：确定性检查、人工评审和模型评分，按任务选择合适的方法。
- 多轮模拟：模拟用户及环境反馈，检查 Agent 在连续交互中的表现。
- 回归评估：修改提示词、模型、工具或流程后，检查已有能力是否退步。

## 2. 与可观测性和训练的关系

[可观测性](../observability/Observability.md) 提供实际运行的记录和问题线索；评估根据场景和标准判断表现。实际失败会话可以转化为评估案例，评估结果也可以帮助确定后续改进方向。

评估不等同于训练。修改提示词或工具后可以进行评估；更新模型参数后同样需要评估。评分时应检查最终产物或外部状态，不能仅依据 Agent 自称完成任务。

## 3. 相关实现

- [DigitalOcean Agent Learning Services](../../ecosystem/digitalocean/AgentLearningServices.md)：通过 Evaluations 与 Simulations 评分和模拟多轮交互，验证修改效果。
- [Arize Phoenix](../../ecosystem/arize/Phoenix.md)：通过数据集、实验、标注与评分比较模型和 Agent 的表现。
