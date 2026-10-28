# 构建引擎启动命令

写入引擎配置后，检查停止信号，再调用 `engine.buildCommand(spec, config)`，生成可执行程序、参数和环境变量。

## 1. 命令内容

根据所选引擎设置模型、推理参数、提示词、输出格式和适用的恢复会话参数。Claude Code 使用其 CLI；Codex 普通执行使用 `codex exec`，支持执行中追加输入的模式使用 `codex app-server`。

配置目录通过 `CLAUDE_CONFIG_DIR` 或 `CODEX_HOME` 传入，适用的认证信息也可通过环境变量传入。此处只组装命令，还未创建引擎进程。

## 2. 存储操作

本步骤在内存中生成命令，不直接读写 MySQL、Redis 或对象存储。

## 3. 下一步

使用选定的可执行程序与环境变量[安装插件](InstallPlugins.md)。
