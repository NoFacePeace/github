# Spec

`spec` 描述 Kubernetes 对象的期望配置，也就是希望 Kubernetes 如何运行或配置该对象。可用的子字段由资源类型决定。

## 1. 在 YAML 中的位置

`spec` 通常是对象的顶层字段，与 `apiVersion`、`kind` 和 `metadata` 同级。下面的 Pod 示例中，`containers` 和 `volumes` 都是 `spec` 下的子字段。示例镜像 `my-backend:1.0` 需替换为实际镜像。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: backend-pod
  labels:
    app: backend
spec:
  containers:
    - name: backend
      image: my-backend:1.0
      volumeMounts:
        - name: cache
          mountPath: /cache
  volumes:
    - name: cache
      emptyDir: {}
```

## 2. 不同资源的配置

| 资源类型 | `spec` 中的字段示例 |
| --- | --- |
| Pod | `spec.containers` 指定运行哪些容器，`spec.volumes` 定义供容器挂载的卷。 |
| Service | `spec.selector` 指定匹配哪些 Pod，`spec.ports` 指定访问端口和后端端口。 |

`metadata` 保存对象的名称、标签等信息。`spec` 记录期望配置，资源支持 `status` 时，`status` 记录 Kubernetes 或控制器观察到的状态。例如，Pod 的 `spec.containers` 指定要运行的容器，`status.containerStatuses` 可显示容器是否就绪。
