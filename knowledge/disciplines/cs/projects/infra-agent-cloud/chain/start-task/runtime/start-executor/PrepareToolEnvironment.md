# 准备 MCP 与 Git 环境

根据本轮执行描述、敏感输入和已准备的路径，处理预安装 MCP 工具及 Git 环境。

## 1. 工具准备

准备适用的 MCP 挂载和配置，取得调整后的敏感输入。这里与 Server 准备工具授权分工不同，主要处理 Runner 启动沙箱需要的本地适配。

## 2. Git 环境

准备任务使用的 Git 配置与环境变量，供后续沙箱执行使用。凭证不应直接混入公开命令参数或普通日志。

### 2.1 准备配置不等于拉取业务仓库

Runner 在宿主侧准备 Git 用户身份、credential helper、命令包装及环境变量，并将配置中的路径转换为沙箱内可访问的路径。这一步为后续 Git 操作提供条件，本身不执行业务仓库的 `git clone`。

需要区分两种场景：

| 场景 | 执行方与时机 | 目的 |
| --- | --- | --- |
| 获取 Agent 配置仓库 | Runner 在启动 Executor 之前执行。 | 获取运行所需的版本文件并生成配置。 |
| 拉取任务需要的业务代码 | Agent 按任务需要，通过终端工具在沙箱内执行 Git。 | 获取需要分析、修改或构建的业务代码。 |

业务代码拉取不是每次任务必经的固定步骤。工作目录已有代码或任务不需要仓库时，可以不执行；具体命令取决于任务指令和工具能力。

### 2.2 Agent 如何拉取业务代码

以通用路径和仓库地址举例，Agent 可通过终端工具执行：

```sh
cd /sandbox/workspace
git clone https://git.example.com/team/project.git
```

调用关系为：

```text
Runner 提前准备 Git 配置、凭证获取方式及工作目录
  ↓
Executor 启动 Agent CLI
  ↓
Agent 根据任务需要调用终端工具
  ↓
终端工具在沙箱内启动 Git 进程
  ├─ 私有 HTTPS 仓库认证时，按 Git 配置调用 credential helper
  ├─ 请求经过沙箱网络规则和 slirp4netns，访问 Git 服务
  └─ 将代码写入 /sandbox/workspace/project
```

这里的 `git.example.com` 和 `/sandbox` 是示例占位。公开仓库可能不需要凭证；实际认证方式与目标服务有关，credential helper 不保证对任意仓库都有效。

业务代码落在沙箱工作目录对应的宿主任务可写目录中。Git 进程与任务其他子进程一样，受网络、seccomp、文件挂载和 cgroup 限制；目标服务必须可达，且需要的凭证已授权。

## 3. 后续步骤

进入[构建 bwrap 隔离](BuildSandbox.md)。
