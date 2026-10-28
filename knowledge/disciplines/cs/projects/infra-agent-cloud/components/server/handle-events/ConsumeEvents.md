# 消费执行事件

Server 启动控制面时，为自身范围的[执行事件 Stream](../../../storage/redis/RunnerEventStream.md)注册消费组处理逻辑。

## 1. 读取

恢复符合条件的未确认消息，再读取新消息。当前 Server 配置的批量大小为 100，实际一次返回数量取决于可读消息。

读取使用消费组，会更新投递及待确认状态；消费组与 Pending 列表属于 Stream 内部元数据，不是独立 key。

## 2. 分组

按回复范围、Runner 范围、Runner、Task、Session、Job 和租约等执行归属字段分组，避免将不同轮次的事件混为一批。

## 3. 后续步骤

进入[校验执行事件](ValidateEvents.md)，判断事件是否属于有效执行以及是否需要继续处理。
