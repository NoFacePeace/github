# docker push

`docker push` 用于将本地镜像上传到镜像仓库，供其他机器拉取和部署。推送需要已有本地镜像，并拥有目标仓库的推送权限。

## 1. 基本用法

```sh
docker push <仓库地址>/<项目>/<镜像>:<标签>
```

尖括号内容是占位符，执行前替换为实际值。镜像名称中不指定仓库主机时，默认使用 Docker Hub；项目或命名空间层级取决于仓库。

## 2. 镜像标签

推送前，本地镜像需要具有目标仓库对应的标签：

```sh
docker tag <源镜像>:<标签> <仓库地址>/<项目>/<镜像>:<标签>
```

`docker tag` 添加本地镜像引用，不会上传镜像。`docker push` 上传镜像，不会启动容器或完成部署。

## 3. 示例

### 3.1 登录并推送镜像

以下以已有本地镜像 `my-app:1.0` 为例。将 `registry.example.com`、`team` 和镜像名称替换为实际值。

```sh
# 登录目标镜像仓库，按提示完成认证
docker login registry.example.com

# 为本地镜像添加目标仓库标签
docker tag my-app:1.0 registry.example.com/team/my-app:1.0

# 上传镜像，需要目标仓库的推送权限
docker push registry.example.com/team/my-app:1.0
```

更多认证方式见 [docker login](Login.md)。

### 3.2 拉取已上传的镜像

其他机器可以在具备访问权限的情况下下载：

```sh
docker pull registry.example.com/team/my-app:1.0
```

[返回 Docker 命令目录](Commands.md)
