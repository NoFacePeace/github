# 读取调度记录

根据当前 Session UUID，通过 Redis `GET` 读取[调度记录](../../../../../storage/redis/dispatch-record/DispatchRecord.md)，再将 JSON 解析为调度数据。

## 1. 读取内容

读取任务及执行标识、调度状态、排队时间、超时截止时间，以及当前阶段保存的加密输入或派发信息。

此时读取记录本身不需要解密执行输入；准备派发时再按需获取并解密。

## 2. 记录不存在

若 key 不存在，结束当前项处理，不继续分配 Runner。不能仅凭等待集合中存在 Session UUID，就假设对应调度记录一定存在。

## 3. 存储操作

这一步只读取 Redis，不新增 Task 或 Session，不删除等待集合成员。
