# 调用链

记录 Agent 云平台从任务发起到结果交付的通用调用流程。

## 1. 发起任务

[发起任务](start-task/StartTask.md)：用户通过首次消息创建新任务，并启动执行。

## 2. 继续任务

[继续任务](continue-task/ContinueTask.md)：用户向已有任务发送消息，补充信息或提出后续要求。

## 3. 停止任务

[停止任务](stop-task/StopTask.md)：取消排队请求或停止正在执行的任务，也支持指定 Session。

## 4. 执行中追加指令

[执行中追加指令](steer-task/SteerTask.md)：在支持的模式下，将新指令交给当前正在运行的引擎。
