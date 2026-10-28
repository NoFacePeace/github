# 启动 Executor

准备流程完成且未收到取消或租约失效信号后，Runtime 使用执行描述和敏感输入启动 Executor。

## 1. 启动边界

Runtime 负责受控目录、隔离和进程生命周期；具体引擎命令、配置及输出解析由 [Executor](../../../../components/Executor.md)负责。本目录按 bwrap 路径展开。

进入 Runtime 的启动接口不代表 Agent 已完成初始化，后续还要根据执行事件判断引擎就绪和实际输出。

## 2. 启动流程

从[准备启动文件与资源限制](PrepareLaunchFiles.md)开始，沿后续链接处理：

```text
准备任务路径、manifest 和隔离辅助文件
  → 创建执行 cgroup
  → 准备 MCP 与 Git 环境
  → 构建 bwrap 参数和网络隔离
  → 创建进程、登记进程信息并传入敏感输入
  → 建立输出、退出与清理监控
```

## 3. 完整启动链

以下按 bwrap 与隔离网络路径描述，程序路径和分组名采用通用示意：

```text
systemd 启动并管理 Runner 服务
  → Runner 调用 runtime.start()
  → 准备文件、cgroup、工具与隔离参数
  → spawn 外层 shell
      → 将自身 PID 写入执行组的 cgroup.procs
      → exec RootlessKit
          → 启动 slirp4netns，建立网络环境
          → 网络脚本设置规则并打开 FD 3
          → seccomp 包装脚本打开 FD 4
          → exec bwrap
              → 设置隔离空间、挂载和环境变量
              → 启动 Node.js Executor
                  → 读取 manifest 与 stdin 敏感输入
                  → 配置工具与引擎
                  → 启动 Agent 引擎
```

这是启动控制链，不是每个箭头都对应一个独立常驻进程。`exec` 替换当前进程程序；RootlessKit、网络辅助进程和沙箱内进程仍组成进程树。

cgroup、隔离与 Executor 进程创建都属于当前 Runtime 启动步骤。Executor 自身的输入解析与引擎初始化，是它开始运行后的后续处理。

## 4. 后续监控

启动后按适用条件开启会话日志备份，并进入输出消费、超时、停止与完成结果监控。执行可在后台继续，Job 消费确认不需要等到整轮完成。
