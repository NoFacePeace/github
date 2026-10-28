# Runner 维护标记

记录某个 Runner 是否被要求停止接收新任务，供执行准入与生命周期管理检查。

## 1. Key 与类型

以下使用通用命名空间：

```text
<namespace>:<runnerScope>:runner:{<runnerScope>}:maintenance:<runnerId>
```

类型为 String，值为序列化 JSON。

## 2. 内容示例

```json
{
  "id": "maintenance-example",
  "requestedAt": "2026-10-28T02:00:00.000Z",
  "reason": "示例：部署升级",
  "origin": "vm"
}
```

以上数据为虚构示例。`id` 是本次维护标识，`origin` 表示 `vm` 或 `admin` 来源。

## 3. 读写时机

- 设置维护：通过 `SET` 保存标记，管理员操作还会校验已有维护的来源。
- 接收 Job：通过 `GET` 检查标记是否存在，存在则阻止正常接单。
- 预留容量：准入脚本也检查该 key，避免继续向维护节点预留执行名额。
- 恢复服务：核对维护身份、心跳与容量条件，通过后删除 key。

## 4. 有效期

不设置 TTL，必须由恢复流程显式清除，避免失败发布或长时间维护期间自动重新接单。
