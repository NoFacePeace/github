# Job Stream

按目标 Runner 保存已投递的执行作业，供节点消费。

## 1. Key 与类型

以下采用通用命名空间：

```text
<namespace>:<runnerScope>:runner:jobs:<runnerId>
```

类型为 Stream。一个 Runner 的 key 包含多条作业消息，不是每个 Job 单独创建 key。

## 2. 消息内容

通过 `XADD` 追加一条消息，字段 `payload` 保存 Job 的 JSON，包括任务与执行标识、目标节点、租约信息及执行清单等。凭证通过授权引用受控领取。

Stream 自动生成的消息 ID 与 Job 中用于发布去重的 `messageId` 是不同标识。

## 3. 写入与保留

投递时使用 `MAXLEN ~` 进行近似长度裁剪，默认上限参数为 10000，可在初始化传输层时调整。本投递路径不为 Stream key 设置 TTL，长度裁剪与 key 过期分别管理。

派发去重检查、追加消息与保存去重标记在同一 Lua 脚本中执行；投递成功不等同于 Runner 已消费。
