# Metadata

`metadata` 是 Kubernetes 资源对象的通用字段。Pod、Service、Deployment 和 Node 都用它保存对象的身份和附加信息。

## 1. 名称与命名空间

- `metadata.name` 是对象的名称。例如，Service `backend` 的名称是 `backend`。
- `metadata.namespace` 表示命名空间级对象所在的命名空间。例如，`default/backend` 中的 `default` 是 Service 所在的命名空间。Node 属于集群级资源，其 `metadata` 中没有 `namespace`。

## 2. 标签与注解

- [Label](Label.md) 对应 `metadata.labels`，用于给对象分类，供选择器匹配和筛选。
- [Annotation](Annotation.md) 对应 `metadata.annotations`，用于记录说明或存放组件约定的配置。
