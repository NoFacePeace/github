# Nemotron 3.5 Super VL

Nemotron 3.5 Super VL 是 NVIDIA 发布的视觉语言模型（VLM），属于[多模态模型](../../models/multimodal/Multimodal.md)，支持视觉理解、文本生成和 Agent 工具调用等任务。

## 1. 模型定位

- 理解图片、网页、PDF 中的文字、表格和图表等视觉信息。
- 结合文本与视觉上下文进行推理，生成文字结果。
- 支持工具使用和多步工作流，可作为 Agent 的模型组件。

模型负责理解输入和生成输出；实际工具执行、状态管理与任务推进仍需应用及运行系统配合。

## 2. 规模与架构

| 项目 | 说明 |
|---|---|
| 总参数量 | 120B |
| 激活参数量 | 约 12B |
| 参数组织 | 混合专家（MoE） |
| 核心架构 | Mamba–Transformer 混合架构 |
| 模态分类 | 视觉语言模型 |

激活参数量描述每次计算使用的参数规模，不等于加载模型所需的内存容量。部署时应结合权重精度、上下文长度和推理引擎计算资源需求。

## 3. 开放与部署

NVIDIA 提供开放权重、相关数据集和训练方法，可通过官方发布页面查找模型下载及 NVIDIA NIM 等部署入口。各项资源的商用、修改和再分发条件应以对应许可证为准。

Super VL 偏向较大规模部署；需要更低内存占用的本地模型时，可了解经过压缩和量化的 Bolt 版本。

## 4. 参考资料

- [官方发布说明](https://blogs.nvidia.com/blog/nemotron-3-5-super-vl/)
- [NVIDIA Nemotron 研究项目](https://research.nvidia.com/labs/nemotron/projects/)
