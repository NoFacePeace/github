# 逐行读取输出

Executor 通过引擎进程句柄取得标准输出、标准错误管道，分别调用 `consumeLines`，并行逐行消费。

## 1. 数据来源

这里读取的是 Claude Code 或 Codex 子进程的输出。Executor 自己的标准输出则用于向 Runtime 发送协议事件，两层管道用途不同。

标准输出的原始行还会保留在内存数组中，供退出后解析最终结果。此步骤不查询 MySQL、Redis 或对象存储。

## 2. 并行控制输入

适用模式下，[处理追加输入](HandleControlInput.md)与输出消费并行进行。

## 3. 下一步

对读取的每一行[记录本地日志](RecordOutputLog.md)。
