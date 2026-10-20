# 参数高效微调

PEFT（Parameter-Efficient Fine-Tuning，参数高效微调）是一类仅训练少量参数来适配已有模型的方法，通常冻结大部分或全部基础模型参数。

## 1. 常见方法

- [LoRA](LoRA.md)：通过低秩矩阵表示权重增量，通常只训练新增的低秩参数。
- Adapter：加入较小的可训练模块。
- Prompt Tuning、Prefix Tuning：训练连续提示向量或前缀表示，不等同于手写提示词。

PEFT 是方法类别，LoRA 是具体方法。参数高效不意味着基础模型无需加载，或所有计算与内存开销都按相同比例减少。

## 2. 同名工具库

Hugging Face PEFT 是实现多种参数高效微调方法的工具库，需要与 PEFT 方法类别区分。

## 3. 参考资料

- [Hugging Face PEFT 文档](https://huggingface.co/docs/peft/index)
