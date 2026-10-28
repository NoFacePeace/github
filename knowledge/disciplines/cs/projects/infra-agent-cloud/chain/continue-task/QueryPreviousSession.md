# 查询上一轮 Session

## 1. 处理内容

从 [Session](../../storage/mysql/session/Session.md) 中按 Task ID 查询，按创建时间倒序取得最近一轮。

上一轮用于后续运行配置兼容性检查、模型变化提示、父 Session 关联及引擎会话恢复标识选择。不存在上一轮时，相关信息允许为空。

## 2. 下一步

进入[解析配置并检查授权](ResolveExecutionConfig.md)。
