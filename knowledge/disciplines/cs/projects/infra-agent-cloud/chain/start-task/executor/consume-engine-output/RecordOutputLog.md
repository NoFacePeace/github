# 记录本地日志

Executor 对标准输出调用适用的 `filterLogLine`，对标准错误调用适用的 `filterErrorLine`，然后交给已打开的 `SessionLog`。

## 1. 写入内容

过滤器可以修改日志行，也可以返回 `null` 排除该行。日志模块忽略空行，按写入队列追加到 `<sessionUuid>.jsonl`；JSON 对象缺少字符串时间戳时补充时间戳。

原始标准输出仍保存在内存中，过滤结果只决定日志内容。文件位置和备份职责见[对象存储 session](../../../../storage/object-storage/Session.md)。

## 2. 存储操作

写 Runner 本地文件，此处不直接写 MySQL、Redis 或对象存储。

## 3. 下一步

等待本行日志写入完成后，[上报原始输出事件](EmitRawOutput.md)。
