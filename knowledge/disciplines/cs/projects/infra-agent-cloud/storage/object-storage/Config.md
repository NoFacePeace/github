# config

保存任务的引擎配置及适用的引擎会话文件备份，是恢复底层 Agent 会话所需的数据来源之一。

## 1. 路径

```text
对象存储：tasks/<taskUuid>/config/
本地目录：<dataRoot>/tasks/<taskUuid>/config/
```

## 2. 内容与用途

当前执行器将 Codex 的 `CODEX_HOME` 和 Claude Code 的 `CLAUDE_CONFIG_DIR` 指向本地 `config/`。因此它不只是静态配置目录，引擎自身的会话数据也由引擎在该根目录下管理。

具体子目录取决于执行引擎。恢复引擎对话还需要有效的会话 ID，平台从 Session 记录中取得 `executor_session_id`，作为后续执行的 `resumeSessionId`；不是仅靠目录名恢复。

## 3. 备份与恢复

常规任务备份会上传本目录，但排除临时文件、缓存、技能及部分插件安装缓存等可重建内容，不能理解为整个目录无差别上传。

继续执行时，本地缺少有效内容才按备份恢复，随后仍会准备版本资源并应用本轮配置。
