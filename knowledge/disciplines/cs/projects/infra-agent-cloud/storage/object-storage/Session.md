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

日志由 Executor 写入 Runner 本地文件，再由 Runtime 上传到对象存储，用于日志留存、查询与恢复相关文件。Server 消费输出事件时保存 Redis 日志，属于另一条保存路径。

### 2.1 执行期间定时备份

远程沙箱 Agent 执行在 Executor 启动、取得执行句柄后，为当前 Session 开启定时器，每 30 秒尝试上传 `<sessionUuid>.jsonl`。首次尝试在开启约 30 秒后；文件不存在、为空或上次上传仍在进行时跳过。

每次上传当前日志文件内容，覆盖同一个 `tasks/<taskUuid>/session/<sessionUuid>.jsonl` 对象。不同 Session 使用不同 key。上传期间日志仍可能增长，定时备份不是逐行实时同步。

该定时器只上传当前日志文件，不上传整个 session 目录或其他目录；执行收尾时关闭定时器并停止适用的进行中上传。

### 2.2 执行结束后的任务备份

Runtime 常规任务备份上传 config 和 session 目录，因此 session 目录中适用的最终输出文件也可随目录备份上传。这是独立于 30 秒日志定时器的流程。

Memory 通过单独的变更监听与备份流程处理；当前代码没有实现 workspace 整体备份。

## 3. 与引擎会话的区别

平台 Session 表示一轮执行，底层引擎会话可以跨多轮执行延续。引擎会话数据由引擎在配置根目录中管理，因此仅恢复本目录中的日志不足以恢复 Agent 对话，更不能恢复旧进程。
