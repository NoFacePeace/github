---
color: "#4A90D9"
---

# Colima

Colima 是用于在 macOS 和 Linux 上运行容器环境的开源工具，基于 Lima 管理 Linux 虚拟机，支持 Docker、containerd 等容器运行时。

## 1. 安装

macOS 可以通过 Homebrew 安装 Colima 和 Docker 命令行客户端：

```sh
brew install colima docker
```

Colima 提供运行环境，Docker 客户端用于执行镜像和容器操作。

## 2. 常用命令

| 命令 | 用途 |
| --- | --- |
| `colima start` | 启动默认容器环境 |
| `colima status` | 查看运行状态 |
| `colima stop` | 停止容器环境 |
| `colima ssh` | 进入 Linux 虚拟机 |
| `colima list` | 查看实例列表 |

### 2.1 colima start

`colima start` 用于创建或启动指定实例的 Linux 虚拟机及容器运行时。

- 首次启动时创建实例，并按需下载虚拟机镜像，需要网络连接。
- 不指定实例名称时使用 `default` 实例，新实例默认使用 Docker 运行时。
- 已有实例会使用保存的配置启动，无需每次重复指定资源参数。
- 实例已运行时，重复执行不会创建另一个实例。调整 CPU、内存等配置时，应先停止实例再启动。

| 参数 | 用途 |
| --- | --- |
| `--cpu <数量>` | 设置虚拟机 CPU 数量 |
| `--memory <GiB>` | 设置虚拟机内存大小 |
| `--disk <GiB>` | 设置虚拟机磁盘容量；已有磁盘能否调整取决于版本，通常不支持缩小 |
| `--runtime docker` | 使用 Docker 运行时 |
| `--runtime containerd` | 使用 containerd 运行时，通常通过 `nerdctl` 操作容器 |
| `--edit` | 启动前通过编辑器修改配置 |

可以使用 `colima start --help` 查看当前版本支持的参数及默认值。

## 3. 示例

### 3.1 启动并验证 Docker 环境

```sh
# 使用默认配置启动，也可恢复已停止的默认实例
colima start

# 使用 Docker 运行时启动默认实例
colima start --runtime docker

# 查看 Colima 状态及当前 Docker context
colima status
docker context show

# 下载并运行测试容器，退出后自动删除容器
docker run --rm hello-world
```

默认实例对应的 Docker context 通常为 `colima`。如果当前 context 指向其他环境，可以执行 `docker context use colima` 后再运行测试容器。

上面的两条启动命令是替代写法；新建默认 Docker 实例时任选一条即可。

### 3.2 运行 Nginx

```sh
# 将本机 8080 端口映射到容器的 80 端口
docker run -d --name colima-nginx -p 127.0.0.1:8080:80 nginx:alpine

# 验证 HTTP 服务
curl http://localhost:8080

# 查看容器日志
docker logs colima-nginx

# 停止并删除示例容器
docker stop colima-nginx
docker rm colima-nginx
```

也可以在浏览器中访问 `http://localhost:8080` 查看 Nginx 默认页面。

### 3.3 停止容器环境

```sh
colima stop
```

停止 Colima 会停止虚拟机，容器服务也随之不可用；再次运行 `colima start` 可启动环境。容器是否自动启动取决于其重启策略。

### 3.4 调整资源后启动

```sh
colima stop
colima start --cpu 4 --memory 8
colima status
```

### 3.5 创建独立实例

```sh
# 创建或启动名为 dev 的实例
colima start dev --cpu 2 --memory 4 --runtime docker

# 查看并停止该实例
colima status dev
colima stop dev
```

命名实例可分别管理配置和运行环境。使用 Docker 时，通过 `docker context ls` 查看对应 context，再用 `docker context use <名称>` 切换目标环境。

## 4. 配置

### 4.1 资源配置

通过 `colima start` 的参数设置 CPU、内存和磁盘容量，配置会保存到对应实例，后续启动可以复用。已有实例调整 CPU、内存前应先停止，完整操作见“3.4 调整资源后启动”。

```sh
# 配置 4 个 CPU、8 GiB 内存并启动
colima start --cpu 4 --memory 8
```

也可以通过 `colima start --edit` 在启动前编辑配置。

### 4.2 Docker 连接配置

Docker 客户端需要连接到 Colima 中的 Docker 服务。默认实例可以通过 Docker context 或 `DOCKER_HOST` 环境变量指定连接地址，使用前先启动 Docker 运行时的 Colima 实例。

#### 4.2.1 Docker context

日常使用 Docker CLI 时，可以切换到默认实例对应的 context：

```sh
docker context use colima
```

#### 4.2.2 DOCKER_HOST

部分开发工具、测试框架和 Docker SDK 会读取 `DOCKER_HOST`，但不读取 Docker context。可以设置环境变量，让这些程序连接 Colima 默认实例：

```sh
export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"
```

| 内容 | 含义 |
| --- | --- |
| `export DOCKER_HOST` | 设置环境变量，当前 Shell 及其启动的子进程可以读取 |
| `unix://` | 通过本机 Unix socket 连接 |
| `${HOME}` | 当前用户的主目录 |
| `.colima/default/docker.sock` | Colima 默认实例的 Docker socket 路径 |

设置后，在当前终端执行 `docker ps`、`docker run` 等命令会连接该地址。此设置只对当前 Shell 会话及其子进程有效；若需要每次打开终端自动设置，可以将上述命令加入 Shell 配置文件，例如 zsh 的 `~/.zshrc`。

`DOCKER_HOST` 会覆盖当前 Docker context 的连接地址；显式指定 `docker --context ...` 或设置 `DOCKER_CONTEXT` 时，context 优先。Docker CLI 已通过 context 正常连接时，通常无需额外设置 `DOCKER_HOST`。

取消当前会话中的设置：

```sh
unset DOCKER_HOST
```

如果已写入 Shell 配置文件，还应删除对应配置行，避免下次打开终端时重新设置。

[返回工具目录](Tools.md)
