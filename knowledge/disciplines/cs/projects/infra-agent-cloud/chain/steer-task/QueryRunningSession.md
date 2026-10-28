# 查询当前执行轮次

## 1. 处理内容

在该 Task 下查询状态为 running 的最新 [Session](../../storage/mysql/session/Session.md)。不存在时返回未投递，由调用方选择后续处理。

Steer 成功时沿用当前 Session，不创建新 Session。仅有 running 状态还不代表引擎控制输入已经就绪。

## 2. 后续流程

正常路径进入[检查租约与执行器就绪](CheckExecutorReady.md)。
