# docker login

`docker login` 用于登录容器镜像仓库，让客户端可以拉取私有镜像或推送有权限的镜像。它与连接本机 Docker 服务是不同的操作，登录成功也不代表拥有所有仓库的访问权限。

## 1. 基本用法

```sh
# 登录默认镜像仓库 Docker Hub
docker login

# 登录指定镜像仓库
docker login <仓库地址>
```

`<仓库地址>` 是占位符，执行前替换为实际仓库主机名及可选端口，不包含 `https://` 或项目路径。按提示完成认证；具体登录方式取决于仓库。

## 2. 常用参数

| 参数 | 完整写法 | 用途 |
| --- | --- | --- |
| `-u <用户名>` | `--username <用户名>` | 指定登录用户名 |
| `-p <密码或令牌>` | `--password <密码或令牌>` | 在命令行中直接传入密码或访问令牌 |
| `--password-stdin` | 同左 | 从标准输入读取密码或访问令牌 |

只指定 `-u` 时，Docker 会提示输入密码。使用 `-p` 可能让凭据出现在 Shell 历史或进程参数中；交互操作可使用密码提示，脚本中优先使用 `--password-stdin`。

## 3. 凭据保存与退出

登录凭据由 Docker 配置或系统凭据助手保存。

```sh
# 退出指定镜像仓库的登录
docker logout <仓库地址>
```

## 4. 示例

### 4.1 登录指定仓库

```sh
docker login registry.example.com
```

将示例域名替换为实际仓库地址。登录后可以执行有权限的镜像拉取或[推送](Push.md)操作。

### 4.2 指定用户名

```sh
docker login registry.example.com -u my-user
```

将 `my-user` 替换为实际用户名，再按提示输入密码或访问令牌。

### 4.3 指定用户名和密码

假设已设置 `USERNAME` 和 `TOKEN` 环境变量：

```sh
docker login registry.example.com -u "$USERNAME" -p "$TOKEN"
```

这里 `-u` 指定用户名，`-p` 传入令牌。即使通过环境变量传值，展开后的凭据仍可能出现在进程参数中。

### 4.4 使用访问令牌登录

假设已设置 `USERNAME` 和 `TOKEN` 环境变量，使用标准输入传入令牌：

```sh
printf '%s' "$TOKEN" |
  docker login registry.example.com -u "$USERNAME" --password-stdin
```

退出该仓库的登录：

```sh
docker logout registry.example.com
```

[返回 Docker 命令目录](Commands.md)
