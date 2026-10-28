# 恢复文件并准备环境

## 1. 处理内容

Runner 消费并校验 Job、回传 accepted，然后准备本轮执行环境。已有引擎恢复标识时按规则恢复任务备份文件；本地目录已有内容时可能跳过恢复，不是每轮都强制重新下载全部文件。

恢复与资源准备共用[准备执行环境](../start-task/runtime/prepare-execution-environment/PrepareExecutionEnvironment.md)，包括凭证、租约续期、Agent 资源和 Memory。对象备份中没有工作目录整体备份，不能承诺恢复所有 workspace 文件。

## 2. 下一步

进入[启动引擎并恢复对话](ResumeEngine.md)。
