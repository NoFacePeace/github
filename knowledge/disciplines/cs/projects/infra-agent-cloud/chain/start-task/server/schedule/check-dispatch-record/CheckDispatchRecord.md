# 检查调度记录

取得当前 Session 的调度处理权后，读取其调度记录并判断是否需要继续派发。

## 1. 处理顺序

```text
读取调度记录
  → 检查调度状态
  → 排队中：同步业务状态并再次确认
  → 检查排队超时
  → 未超时：继续准备分配 Runner
```

- [读取调度记录](ReadDispatchRecord.md)：根据 Session UUID 获取已保存的调度信息。
- [检查调度状态](check-dispatch-state/CheckDispatchState.md)：区分排队请求与其他状态，避免按新请求重复派发。
- [检查排队超时](CheckQueueTimeout.md)：根据截止时间判断是否标记为失败。

## 2. 处理边界

本阶段仍由持有调度 claim 的 Server 执行。读取记录不表示已经获得 Runner 执行租约，也不代表 Agent 已启动。
