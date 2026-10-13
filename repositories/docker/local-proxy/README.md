# local-proxy

基于 Ubuntu 24.04 和 Squid 的本地 HTTP/HTTPS 代理。Squid 使用 `-N` 在前台运行，由 Docker 管理进程。

## 1. 运行信息

- 容器名称：`local-proxy`
- 镜像名称：`local/local-proxy`
- 本地代理地址：`http://127.0.0.1:3128`
- 端口映射：`127.0.0.1:3128:3128`
- HTTP 允许目标端口 80、443，HTTPS CONNECT 仅允许目标端口 443。
- 禁用缓存；当前配置不要求代理认证，创建容器时应保留本地回环地址绑定。

## 2. 启动故障记录

2026-10-13，启动已有容器后，容器立即以退出码 `1` 退出。日志如下：

```text
FATAL: Squid is already running: Found fresh instance PID file (/tmp/squid.pid) with PID 1
```

容器此前停止后，`/tmp/squid.pid` 残留，记录的 PID 为 `1`。再次启动时，前台运行的 Squid 同样使用容器内 PID `1`，因而将残留记录误判为已有实例，拒绝启动。

尝试替换残留 PID 文件后，又出现文件权限错误：

```text
FATAL: failed to open /tmp/squid.pid: (13) Permission denied
```

## 3. 修复方式与持久化

本次将**现有容器内** `/etc/squid/squid.conf` 的 PID 配置改为：

```conf
pid_filename none
```

随后重新启动，确认容器处于运行状态，端口映射为 `127.0.0.1:3128->3128/tcp`。本次确认了容器运行状态，未验证外部网站的代理访问。

当前 Squid 使用 `-N` 前台运行，Docker 已负责进程生命周期管理，因此可以禁用 PID 文件，避免残留文件阻碍重启。

**仓库中的 `squid.conf` 仍为 `pid_filename /tmp/squid.pid`，本次仅记录故障，没有修改源配置。** 为使重新构建、创建的容器也包含修复，需要先将仓库中的这一行改为 `pid_filename none`，再构建镜像并重新创建容器。仅构建镜像不会更新已有容器。

## 4. 构建与首次创建

在本目录执行：

```sh
docker build -t local/local-proxy .
docker run -d --name local-proxy \
  -p 127.0.0.1:3128:3128 \
  local/local-proxy
```

`docker run` 适用于尚无同名容器的情况；已有容器使用下一节的启动命令。

## 5. 常用命令

```sh
# 启动或停止已有容器
docker start local-proxy
docker stop local-proxy

# 查看状态和启动日志
docker ps -a --filter name=^/local-proxy$
docker logs --tail 50 local-proxy

# 查看 Squid 写入文件的运行日志
docker exec local-proxy tail -n 50 /var/log/squid/cache.log

# 验证 HTTPS 代理访问
curl --proxy http://127.0.0.1:3128 --connect-timeout 10 --max-time 30 -I https://example.com
```

`docker logs` 中会保留此前启动失败的记录，判断当前状态时应结合日志时间和 `docker ps` 的结果。
