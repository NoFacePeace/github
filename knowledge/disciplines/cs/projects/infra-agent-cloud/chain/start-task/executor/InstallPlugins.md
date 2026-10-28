# 安装插件

Executor 根据执行描述中的 `installedPlugins`，在当前执行环境内调用引擎原生 CLI 安装插件；没有插件时跳过。

## 1. 安装流程

1. 将插件路径解析到任务配置目录下。
2. 读取对应引擎的 marketplace 元数据，校验名称。
3. 调用 `plugin marketplace add <directory>` 注册来源；同一目录在本次安装中只注册一次。
4. Claude Code 调用 `plugin install <name>@<marketplace>`；Codex 调用 `plugin add <name>@<marketplace>`。

单次安装命令超时为 60 秒。安装失败会抛出错误，进入执行失败处理；安装后再次检查停止信号。

## 2. 存储操作

读取本地插件元数据，安装命令可能写入引擎管理的本地配置与插件文件。此步骤不直接读写 MySQL、Redis 或对象存储。配置目录职责见[对象存储 config](../../../storage/object-storage/Config.md)。

## 3. 下一步

插件准备完成后[打开执行日志](OpenExecutionLog.md)。
