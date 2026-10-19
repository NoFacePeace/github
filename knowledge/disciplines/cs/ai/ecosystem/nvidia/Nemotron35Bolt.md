# Nemotron 3.5 Bolt

Nemotron 3.5 Bolt 是 NVIDIA 面向本地 Agent 场景的模型。本文主要介绍 Bolt 74B-A9B 24GB：它是从 Super VL 压缩得到的视觉语言模型（VLM），属于[多模态模型](../../models/multimodal/Multimodal.md)，支持文字、图像和视频输入以及文字输出。

## 1. 模型定位

- 面向本地运行的理解、推理、工具调用和多步任务。
- 通过压缩和量化降低内存占用，兼顾能力与推理速度。
- 提供完整精度和量化权重，用于不同的定制与部署需求。

## 2. 规模与架构

| 项目 | 说明 |
|---|---|
| 总参数量 | 74B |
| 激活参数量 | 约 9B |
| 参数组织 | 混合专家（MoE） |
| 核心架构 | Mamba–Transformer 混合架构 |
| 本地量化版 | 以约 24GB 内存预算为目标，主要专家权重采用约 2-bit 量化 |
| 完整精度版 | BF16，主要面向定制与后训练，内存需求明显高于量化版 |

“24GB”是特定量化版本的部署目标，不代表 BF16 权重大小，也不意味着任意 24GB 显存设备都能运行所有配置。

## 3. 本地部署

官方重点优化 NVIDIA RTX Spark 与 DGX Spark，提供 llama.cpp 的 GGUF 权重及 vLLM 部署方式。上下文缓存、视觉组件、辅助模型和运行程序也会占用内存，应核对具体部署配置及引擎支持。

对于 16GB 统一内存的电脑，74B-A9B 24GB 版本不适合作为实用的本地部署选择。可以选择更小的模型，并控制上下文长度。

## 4. 开放状态

74B-A9B 24GB 版本已开放权重，包括 GGUF 版本；也提供 BF16 权重及辅助解码模型。使用、修改和再分发应查看所下载资源的具体许可证。

截至 2026 年 10 月 19 日，官方另预告 Bolt E4B 4GB 将于 10 月 23 日发布。其定位是更轻量的对话与物理 Agent，不能直接套用 74B-A9B 的规格和部署结论。

## 5. 参考资料

- [官方技术介绍](https://research.nvidia.com/labs/nemotron/Nemotron-3.5-Bolt-24GB/)
- [24GB 模型权重与说明](https://huggingface.co/nvidia/NVIDIA-Nemotron-3.5-Bolt-74B-A9B-24GB)
- [GGUF 权重与运行方式](https://huggingface.co/nvidia/NVIDIA-Nemotron-3.5-Bolt-74B-A9B-24GB-GGUF)
- [BF16 权重与说明](https://huggingface.co/nvidia/NVIDIA-Nemotron-3.5-Bolt-74B-A9B-BF16)
- [Super VL 与 Bolt 发布说明](https://blogs.nvidia.com/blog/nemotron-3-5-super-vl/)
