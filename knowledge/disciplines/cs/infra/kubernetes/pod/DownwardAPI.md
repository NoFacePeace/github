# Downward API

Downward API 让容器中的应用通过环境变量或挂载文件读取所在 Pod 的信息和容器的资源配置。Kubelet 负责把这些值提供给容器，应用可以直接从进程环境或文件中读取。字段含义见 [Metadata](../resource/metadata/Metadata.md)。

## 1. 通过环境变量读取 Pod 字段

在 `backend-pod` 的 `backend` 容器配置中，也就是 `spec.containers` 对应条目的 `env` 下，用 `valueFrom.fieldRef` 引用 Pod 字段。下面的例子读取 Pod 名称、单个标签和单个注解。

```yaml
env:
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: POD_APP
    valueFrom:
      fieldRef:
        fieldPath: "metadata.labels['app']"
  - name: POD_DESCRIPTION
    valueFrom:
      fieldRef:
        fieldPath: "metadata.annotations['example.com/description']"
```

容器启动后，`POD_NAME` 的值是 `backend-pod`。标签和注解取自 `backend-pod` 自身的 `metadata`。环境变量在容器启动时确定，后续对 Pod 元数据的修改不会刷新运行中容器的环境变量。

## 2. 通过挂载文件读取整组标签和注解

在 Pod 的 `spec.volumes` 定义 `downwardAPI` 卷，再在容器的 `volumeMounts` 中挂载。下面是 Pod `spec` 的片段，`my-backend:1.0` 需要替换为实际镜像。

```yaml
volumes:
  - name: pod-info
    downwardAPI:
      items:
        - path: labels
          fieldRef:
            fieldPath: metadata.labels
        - path: annotations
          fieldRef:
            fieldPath: metadata.annotations
containers:
  - name: backend
    image: my-backend:1.0
    volumeMounts:
      - name: pod-info
        mountPath: /etc/pod-info
        readOnly: true
```

应用可以分别读取 `/etc/pod-info/labels` 和 `/etc/pod-info/annotations`。挂载整个卷时，标签或注解的变更会在 kubelet 同步后反映到文件中。使用 `subPath` 挂载单个文件时，该文件不会收到后续更新。

## 3. 读取容器的 CPU 和内存配置

`resourceFieldRef` 可读取容器的 CPU 和内存请求及限制。下面的 `env` 片段读取 `backend` 容器的 CPU 请求和内存限制。`divisor` 指定输出单位。若 CPU 请求为 `250m`、内存限制为 `256Mi`，输出分别为 `250` 和 `256`。

```yaml
env:
  - name: CPU_REQUEST_MILLICORES
    valueFrom:
      resourceFieldRef:
        containerName: backend
        resource: requests.cpu
        divisor: 1m
  - name: MEMORY_LIMIT_MIB
    valueFrom:
      resourceFieldRef:
        containerName: backend
        resource: limits.memory
        divisor: 1Mi
```
