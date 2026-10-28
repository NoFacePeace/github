# 回传结果与执行收尾

## 1. 处理内容

执行期间经 Runtime 持续回传输出与进度，由 [Server 事件处理](../../components/server/handle-events/HandleEvents.md)保存适用的日志、回放和业务状态。

退出后进行 [Runtime 执行收尾](../start-task/runtime/finalize-execution/FinalizeExecution.md)：停止定时日志备份、清理环境、登记适用的后台 Memory 与 config/session 备份，再发布最终结果。Server 据此处理新 Session 和 Task 状态。

发布结果不等于后台备份已经上传完成。以后再次继续时，还会创建下一轮 Session。
