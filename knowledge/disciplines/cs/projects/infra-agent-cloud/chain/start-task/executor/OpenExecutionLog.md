# 打开执行日志

Executor 调用 `SessionLog.open(spec)`，为本轮平台 Session 打开日志文件。

## 1. 文件位置

在执行描述指定的 `session/` 目录创建或打开 `<sessionUuid>.jsonl`，以追加模式写入；新文件创建权限为 `0600`。目录不存在时递归创建。

日志逐行排队写入，忽略空行以及被过滤器排除的内容。JSON 对象没有字符串类型的时间戳时，会补充时间戳。

## 2. 存储操作

本步骤操作 Runner 本地文件，不直接写 MySQL、Redis 或对象存储。文件的备份职责见[对象存储 session](../../../storage/object-storage/Session.md)。

## 3. 下一步

日志句柄就绪后[启动引擎进程](SpawnEngine.md)。
