# Label

Label（标签）是写在 `metadata.labels` 下的字符串键值对，用于标记对象，供标签选择器匹配。

## 1. 给 Pod 添加标签

下面是 `backend-pod` 的元数据片段。`app: backend` 标记这个 Pod 属于 `backend` 应用。

```yaml
metadata:
  name: backend-pod
  labels:
    app: backend
```

## 2. 按标签选择对象

Service 的 `spec.selector` 写成 `app: backend` 时，会匹配同一命名空间里带有这个标签的 Pod。写在 Service 自身 `metadata.labels` 中的标签只标记该 Service 对象。

也可以用标签选择器查找 `default` 命名空间内带有该标签的 Pod。

```bash
kubectl get pods -n default -l app=backend
```
