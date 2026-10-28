# session

保存平台针对各轮执行记录的日志与输出，不是底层 Agent 引擎的配置根目录。

## 1. 路径

```text
对象存储：tasks/<taskUuid>/session/
本地目录：<dataRoot>/tasks/<taskUuid>/session/
```

同一任务目录下可保存不同平台 Session 的文件，例如：

```text
<sessionUuid>.jsonl
last-message-<sessionUuid>.txt
```

前者是执行器逐行记录的日志，后者是适用引擎保存的最终输出。具体文件取决于执行路径。

## 2. 备份与用途

Runtime 常规任务备份会上传本目录，适用执行路径还会持续备份会话日志。它用于日志留存、查询与恢复相关文件。

## 3. 与引擎会话的区别

平台 Session 表示一轮执行，底层引擎会话可以跨多轮执行延续。引擎会话数据由引擎在配置根目录中管理，因此仅恢复本目录中的日志不足以恢复 Agent 对话，更不能恢复旧进程。
