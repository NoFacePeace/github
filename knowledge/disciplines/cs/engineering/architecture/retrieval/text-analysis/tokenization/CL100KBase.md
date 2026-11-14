# cl100k_base

cl100k_base 是 OpenAI 提供的一套基于字节级 BPE 的 token 编码方案，可通过开源分词库 tiktoken 加载。它规定文本如何预切分、字节片段如何合并，以及最终 token 与整数 ID 如何对应。

## 1. 与算法和工具的关系

| 名称 | 类型 | 职责 |
| --- | --- | --- |
| BPE | 算法方法 | 通过相邻单元合并构建和使用子词词表 |
| cl100k_base | 具体编码方案 | 提供固定的预切分规则、合并优先级、ID 映射及特殊 token 配置 |
| tiktoken | 分词工具库 | 加载编码方案，执行编码、解码与 token 计数 |

模型接收的是 token ID 序列。这里的 token 可能是完整词、词的一部分、带空格的片段，或不足一个完整字符的字节片段，切分结果不一定符合中文词边界。

## 2. 使用方式

安装 tiktoken 后，可以在 Python 中按编码名称加载：

```python
import tiktoken

encoding = tiktoken.get_encoding("cl100k_base")
text = "苹果手机"
token_ids = encoding.encode(text)

print(token_ids)
print(len(token_ids))
print(encoding.decode(token_ids))
print([encoding.decode_single_token_bytes(i) for i in token_ids])
```

`encode()` 返回 token ID 列表，列表长度是该字符串在此编码下的 token 数量。`decode()` 将完整序列还原为文本；查看单个 token 的字节内容时，可使用 `decode_single_token_bytes()`，避免把不完整的 UTF-8 字节片段误当成独立字符。

首次加载可能需要联网下载编码数据，缓存后可以在本地使用。对普通字符串进行编码和计数无需调用模型 API，也无需模型权重。

## 3. 与模型的配合

不同模型可能使用不同编码。OpenAI 官方文档说明，text-embedding-3-small 等第三代嵌入模型使用 cl100k_base。已知目标模型时，也可以通过 `tiktoken.encoding_for_model(model_name)` 获取工具库中对应的编码。

token ID 的含义依赖具体编码，不能把某个编码下的 ID 序列直接交给采用另一编码的模型。token 数量也不是固定的字数或单词数，不同编码可能产生不同结果。

编码名称本身不表示模型的上下文窗口。模型支持多长输入，应查看相应模型的限制；普通字符串计数也不能直接代替包含消息格式、工具定义等内容的完整 API 请求用量。

## 4. 在检索系统中的用途

生成嵌入或准备模型上下文时，可以使用与目标模型匹配的编码统计长度，辅助分块和控制输入规模。按 token 数切块时，还需处理完整文本的解码和语义边界，避免直接在不完整字符处截断。

全文检索的倒排索引可以继续使用面向词项匹配的分析方案。模型编码和检索分词的目标不同，是否共用切分结果需要通过实际查询验证。

## 5. 参考资料

- [OpenAI Cookbook：使用 tiktoken 统计 token](https://developers.openai.com/cookbook/examples/how_to_count_tokens_with_tiktoken)：编码、解码与不同编码的差异；该示例已归档，模型对应关系应结合当前文档确认。
- [OpenAI：向量嵌入指南](https://developers.openai.com/api/docs/guides/embeddings)：嵌入模型的编码选择与 token 计数。
