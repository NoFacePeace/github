# DOCKER_HOST

`DOCKER_HOST` 用于指定 Docker 客户端连接的 Docker 服务地址。部分开发工具、测试框架和 Docker SDK 也会读取该变量。

## 1. 设置

```sh
export DOCKER_HOST="<连接地址>"
```

`<连接地址>` 是占位符，执行前替换为实际地址。常见连接形式如下：

| 形式 | 用途 |
| --- | --- |
| `unix:///路径/docker.sock` | 通过本机 Unix socket 连接 |
| `ssh://用户@主机` | 通过 SSH 连接远程 Docker 服务，需要相应 SSH 与 Docker 访问权限 |
| `tcp://主机:端口` | 通过 TCP 连接，服务端需启用相应监听；TLS 认证需要额外配置 |

变量只指定连接地址，不会启动 Docker 服务，也不会登录镜像仓库。

## 2. 与 Docker context 的关系

`DOCKER_HOST` 会覆盖当前通过 `docker context use` 选择的 context 的连接地址；显式指定 `docker --context ...` 或设置 `DOCKER_CONTEXT` 时，context 优先。

Docker CLI 已通过 context 正常连接时，通常无需额外设置 `DOCKER_HOST`。部分工具不读取 Docker context，但读取该环境变量，此时可以用它指定目标服务；具体支持情况以工具文档为准。

## 3. 生效范围与取消

`export` 设置的变量对当前 Shell 会话及其子进程有效。需要在每次打开终端时自动设置，可以将命令加入对应的 Shell 配置文件，例如 zsh 的 `~/.zshrc`。

取消当前会话中的设置：

```sh
unset DOCKER_HOST
```

如果已写入 Shell 配置文件，还应删除对应配置行，避免下次打开终端时重新设置。

## 4. 示例

### 4.1 连接 Colima 默认实例

先启动使用 Docker 运行时的 [Colima](../Colima.md) 默认实例，再设置连接地址：

```sh
colima start --runtime docker
export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"

# 查看目标服务上的容器
docker ps
```

| 内容 | 含义 |
| --- | --- |
| `export DOCKER_HOST` | 设置环境变量，供当前 Shell 及其子进程读取 |
| `unix://` | 通过本机 Unix socket 连接 |
| `${HOME}` | 当前用户的主目录 |
| `.colima/default/docker.sock` | Colima 默认实例的 Docker socket 路径 |

设置后，读取该变量的 `docker ps`、`docker run` 等命令会连接 Colima 默认实例。如果已设置 `DOCKER_CONTEXT`，需要先取消它才能让 Docker CLI 使用此地址。

[返回环境变量目录](Env.md)
