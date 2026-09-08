# Kubernetes 安全

## 1. 认证与授权

客户端访问 Kubernetes API Server 时，需要先确定请求者身份，再检查其是否有权限执行操作。

- **认证**：确认“你是谁”，例如验证客户端证书、外部身份平台的令牌或 ServiceAccount Token。
- **授权**：判断“你能做什么”，RBAC 是 Kubernetes 支持的一种授权机制。

认证成功不代表有权访问所有资源。这里控制的是 Kubernetes API 的访问，不会自动实现业务应用内部的用户权限。

## 2. RBAC

### 2.1 概念对应

RBAC（Role-Based Access Control）通过角色组织权限，再通过绑定将角色授予主体。

```text
主体 → RoleBinding / ClusterRoleBinding → Role / ClusterRole → 权限规则
```

| RBAC 概念 | Kubernetes 对应 | 示例 |
| --- | --- | --- |
| 主体 | User、Group、ServiceAccount | 用户 alice、开发组、程序服务账号 |
| 角色 | Role、ClusterRole | Pod 只读角色 |
| 角色分配 | RoleBinding、ClusterRoleBinding | 将 Pod 只读角色绑定给 alice |
| 权限 | 角色中的 rules | 允许读取 Pod |
| 资源 | resources | pods、deployments、secrets |
| 操作 | verbs | get、list、watch、create、delete |

### 2.2 主体

| 主体 | 含义 | 管理方式 |
| --- | --- | --- |
| User | 单个用户身份，也可以代表外部程序身份 | 通常由外部认证系统提供 |
| Group | 一组用户身份，用于批量授权 | 组成员关系来自认证结果 |
| ServiceAccount | 通常供 Pod 中的程序使用的服务账号 | Kubernetes API 资源，属于指定命名空间 |

Kubernetes 没有可直接创建的普通 User、Group API 对象。RoleBinding 中写入用户名，并不等于创建用户；请求必须先通过认证，得到匹配的身份。

ServiceAccount 的身份名称形如 `system:serviceaccount:dev:report-reader`，其中 `dev` 是命名空间，`report-reader` 是账号名称。创建服务账号本身不意味着自动获得业务所需权限。

### 2.3 角色与绑定

| 资源类型 | 作用 |
| --- | --- |
| Role | 属于一个命名空间，定义该命名空间内的资源权限 |
| ClusterRole | 不属于命名空间，可定义集群级资源权限，也可定义可复用的命名空间资源权限 |
| RoleBinding | 在所在命名空间内，将 Role 或 ClusterRole 的适用权限授予主体 |
| ClusterRoleBinding | 在集群范围将 ClusterRole 的权限授予主体 |

ClusterRole 不一定意味着全局授权，最终范围取决于绑定方式。例如，通过 `dev` 中的 RoleBinding 引用 Pod 只读 ClusterRole，只授予 `dev` 中的 Pod 读取权限。

RoleBinding 引用 Role 时，该 Role 必须位于同一命名空间；ClusterRoleBinding 只能引用 ClusterRole。RoleBinding 不能授予 Node 等集群级资源的访问权限。

### 2.4 权限规则

```yaml
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
```

- `apiGroups`：资源所属 API 组，空字符串表示核心组；Deployment 属于 `apps` 组。
- `resources`：资源类型，使用 API 中的资源名；子资源单独表示，例如 `pods/log`。
- `verbs`：允许的操作，`get` 读取单个对象，`list` 列出对象，`watch` 监听变化。
- `resourceNames`：可选，用于限制某些操作可访问的对象名称；不能用它限制顶层 `create` 或 `deletecollection`。限制 `list`、`watch` 时，客户端需携带匹配的名称字段选择器。

Kubernetes RBAC 权限是累加的，没有显式拒绝规则。一个角色未授予删除权限，不代表其他角色不能授予；应检查主体所有相关绑定。

### 2.5 示例：允许用户读取 dev 中的 Pod

以下配置假设 `dev` 命名空间已经存在，且认证后的用户名为 `alice`：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
  namespace: dev
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: alice-pod-reader
  namespace: dev
subjects:
  - kind: User
    name: alice
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io
```

请求检查过程：

1. API Server 完成认证，确定请求者是 `alice`。
2. RBAC 根据主体找到匹配的绑定，再读取关联角色的规则。
3. 检查请求的命名空间、API 组、资源和操作是否匹配。
4. 读取 `dev` 中的 Pod 可以匹配上述规则；删除 Pod 则不匹配。若没有其他授权规则允许，删除请求会被拒绝。

可以用当前身份检查权限：

```bash
kubectl auth can-i list pods -n dev
kubectl auth can-i delete pods -n dev
```

这些命令只查询权限，不执行列出或删除操作。结果取决于当前身份的全部授权配置。

## 3. 参考资料

- [Kubernetes 官方文档：认证](https://kubernetes.io/docs/reference/access-authn-authz/authentication/)
- [Kubernetes 官方文档：RBAC 授权](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
