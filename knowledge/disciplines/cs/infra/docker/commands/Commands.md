---
color: "#5B6CDA"
---

# Docker 命令

记录 Docker 客户端的常用命令。执行前需要启动 Docker 服务；使用 Colima 时，可以先执行 `colima start`。

## 1. 环境与连接

| 命令 | 用途 |
| --- | --- |
| `docker version` | 查看客户端与服务端版本 |
| `docker info` | 查看 Docker 服务信息 |
| `docker context ls` | 查看连接环境列表 |
| `docker context show` | 查看当前 context |
| `docker context use colima` | 切换到 Colima 默认实例的 context |

## 2. 镜像

| 命令 | 用途 |
| --- | --- |
| `docker image ls` | 查看本地镜像 |
| `docker pull <镜像>:<标签>` | 下载镜像 |
| `docker build -t <镜像>:<标签> .` | 使用当前目录作为构建上下文构建镜像 |
| `docker tag <源镜像>:<标签> <目标镜像>:<标签>` | 为本地镜像添加新标签，例如指定目标仓库路径 |
| [docker push](Push.md) `<镜像>:<标签>` | 将本地镜像上传到镜像仓库 |
| `docker image rm <镜像>` | 删除本地镜像 |

## 3. 镜像仓库登录

| 命令 | 用途 |
| --- | --- |
| [docker login](Login.md) | 登录默认镜像仓库 Docker Hub |
| `docker login <仓库地址>` | 登录指定镜像仓库 |
| `docker logout <仓库地址>` | 退出指定镜像仓库的登录 |

认证、凭据保存和访问令牌示例见 [docker login](Login.md)。

## 4. 容器

| 命令 | 用途 |
| --- | --- |
| `docker ps` | 查看运行中的容器 |
| `docker ps -a` | 查看全部容器，包括已停止的容器 |
| `docker run --rm <镜像>` | 创建并运行容器，退出后自动删除容器 |
| `docker run -d --name <名称> <镜像>` | 创建并在后台运行容器 |
| `docker stop <容器>` | 请求停止容器，超时后强制终止 |
| `docker start <容器>` | 启动已存在的停止状态容器 |
| `docker restart <容器>` | 重启容器 |
| `docker rm <容器>` | 删除已停止的容器 |

`docker run` 创建新容器，`docker start` 启动已有容器。`<容器>` 可以是容器名称或 ID；尖括号内容是占位符，执行前需替换。

## 5. 日志与排查

| 命令 | 用途 |
| --- | --- |
| `docker logs <容器>` | 查看容器日志 |
| `docker logs -f --tail 100 <容器>` | 查看最后 100 行日志并持续跟踪 |
| `docker exec -it <容器> sh` | 在运行中的容器内启动交互式 Shell，需要镜像包含 `sh` |
| `docker inspect <容器>` | 查看容器配置与状态详情 |
| `docker stats` | 查看运行中容器的资源使用情况 |

## 6. 示例

### 6.1 运行 Nginx 并查看日志

```sh
# 创建容器，仅通过本机 8080 端口提供服务
docker run -d --name demo-nginx -p 127.0.0.1:8080:80 nginx:alpine

# 查看运行状态及日志
docker ps
docker logs demo-nginx

# 验证 HTTP 服务
curl http://localhost:8080

# 停止并删除示例容器
docker stop demo-nginx
docker rm demo-nginx
```

### 6.2 镜像仓库操作

- [登录仓库与访问令牌认证示例](Login.md#4-示例)
- [登录并推送镜像示例](Push.md#31-登录并推送镜像)

[返回 Docker 目录](../Docker.md)
