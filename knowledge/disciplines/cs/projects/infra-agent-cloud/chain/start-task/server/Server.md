# Server

发起任务时，Server 负责接收请求、准备配置、保存业务记录，以及将执行请求交给 Runner。

## 1. 提交

[提交](submit/Submit.md) 整理从查询 Agent、创建任务记录到加入调度队列的步骤。

## 2. 调度

[调度](schedule/Schedule.md) 整理 Server 内部调度器从等待队列选取请求、分配 Runner 并下发作业的过程。
