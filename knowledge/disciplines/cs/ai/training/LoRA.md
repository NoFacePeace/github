# LoRA

LoRA（Low-Rank Adaptation，低秩适配）是一种参数高效微调方法，通常冻结基础模型权重，用少量新增参数学习模型的调整量。

## 1. 基本原理

对于权重矩阵 W，用两个较小的矩阵表示权重增量：

```text
W' = W + s × B × A
```

W 保持冻结，A 和 B 参与训练，s 为缩放系数。若 W 的形状为 d_out × d_in，可令 A 为 r × d_in，B 为 d_out × r，其中 r 通常远小于原矩阵维度，以减少可训练参数量。

## 2. 使用流程

1. 选择基础模型和训练数据，配置目标层、秩及训练参数。
2. 冻结基础权重，训练并保存 LoRA 参数。
3. 推理时加载兼容的基础模型与 LoRA；支持时也可将权重增量合并到基础模型。

LoRA 可用于语言模型的领域适配，也可用于图像生成模型的画风定制。LoRA 文件通常不是完整模型，需要配套的基础模型。

## 3. 参考资料

- [LoRA 原始论文](https://arxiv.org/abs/2106.09685)
- [Hugging Face PEFT：LoRA](https://huggingface.co/docs/peft/conceptual_guides/lora)
