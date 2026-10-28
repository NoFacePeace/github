# 停止任务

用户请求终止已有 Task 的运行轮次。按 Task 停止会处理其 running Session；也支持指定单个 Session。本文以远程沙箱执行分支为主。

## 1. 入口与流程

从[接收停止请求](ReceiveStopRequest.md)开始，按各步骤的分支链接阅读：

```text
接收停止请求并校验权限
  → 查询 Task 与 running Session
  → 事务写入停止请求标记
  → 尝试取消排队
      ├─ 成功：处理取消结果 → Session 进入 stopped
      └─ 未取消：检查执行租约
          ├─ 暂无租约：保留停止意图，等待后续重试
          └─ 有租约：定位 Runner → 发布 stop 命令
              → Runner 校验命令与租约
              → 中止准备或停止执行进程
              → 执行收尾并回传终态
              → Server 保存 Session 停止终态
```

接口侧在提交目标轮次停止请求后，还会更新 Task 状态并返回；这条路径不等待完整的异步停止链结束。

## 2. 三个不同的时刻

| 时刻 | 含义 |
| --- | --- |
| 点击 Stop | 发出停止请求 |
| 请求成功、Task 显示 stopped | 不保证目标 Session 已进入终态，也不保证进程已退出 |
| Session 实际变为 stopped | 本轮业务状态结束；没有其他 running Session 时可进一步检查 Continue 条件 |

## 3. 与继续任务的关系

Stop 不创建新 Session。停止后再发送消息继续任务，才会创建新的平台 Session，适用时恢复原引擎对话。任务产物、配置与授权仍需满足继续执行的要求。
