# KUBE_EDITOR

`KUBE_EDITOR` 是指定 `kubectl edit` 所用编辑器的环境变量。

## 1. 设置

`KUBE_EDITOR` 环境变量用于指定 `kubectl edit` 使用的编辑器。

```sh
export KUBE_EDITOR=/usr/bin/vim
```

| 内容 | 含义 |
| --- | --- |
| `export KUBE_EDITOR` | 设置环境变量，供当前 Shell 及其子进程读取 |
| `/usr/bin/vim` | Vim 可执行程序的路径，需要该程序存在 |

未设置 `KUBE_EDITOR` 时，kubectl 会尝试使用 `EDITOR`，再使用平台默认编辑器。此变量只改变本地编辑器选择，不会修改集群配置。

该设置只对当前 Shell 会话及其子进程有效。需要在每次打开终端时自动设置，可以将命令加入对应的 Shell 配置文件，例如 zsh 的 `~/.zshrc`。

取消当前会话中的设置：

```sh
unset KUBE_EDITOR
```

## 2. 示例

### 2.1 使用 Vim 编辑 Deployment

```sh
export KUBE_EDITOR=/usr/bin/vim
kubectl edit deployment my-app -n default
```

将 `my-app` 和 `default` 替换为实际资源名称及命名空间。执行前确认当前集群上下文，可以通过 `kubectl config current-context` 查看。

kubectl 会获取资源配置并用 Vim 打开：

- 按 `i` 进入插入模式，修改内容。
- 按 Esc 后输入 `:wq`，保存退出；kubectl 会尝试将修改提交到集群。
- 按 Esc 后输入 `:q!`，放弃尚未保存的修改并退出。如果此前已保存，退出不会撤销已保存的内容。

[返回环境变量目录](EnvironmentVariables.md)
