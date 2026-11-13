# 手动上下文压缩完整调用链

本文记录 Codex 原仓库从提交手动压缩操作到任务结束的完整主链。当前 Go 项目尚未实现该流程，文中的 Rust 方法用于后续设计参考。

## 1. 流程概览

```text
提交 Op::Compact
  → 操作入队
  → Session 接收并分发
  → 替换当前任务，创建独立 Turn
  → 启动 CompactTask
  → 选择并执行压缩策略
  → 应用新上下文、更新用量
  → 任务收尾，通过事件流报告终态
```

手动压缩是独立工作流程，完成后不会自动恢复被替换的普通任务。客户端是否允许执行中发起手动压缩，由界面或协议入口决定；本文描述 Core 接收到操作后的行为。

## 2. 提交操作与调度任务

```text
用户执行 /compact 或客户端发起压缩请求
  → 调用方转换为 Op::Compact
  → CodexThread.submit(Op::Compact)
    → CodexThread.submit_with_trace()
      → SessionIo.submit_with_trace()
        → SessionIo.submit_with_id()
          → 将 Submission 写入输入通道
              ↓
Session 后台 submission_loop()
  → 匹配 Op::Compact
  → handlers::compact()
    → abort_all_tasks(TurnAbortReason::Replaced)
      → 取消并清理被替换的任务
    → new_turn_with_default_settings()
      → 创建独立 TurnContext
    → spawn_task(..., CompactTask)
      → 确保旧任务已停止，清理连接器选择
      → start_task()
        → 注册 ActiveTurn / RunningTask
        → 准备取消信号与任务上下文
        → tokio::spawn()：启动后台异步任务
          → CompactTask.run()
            → emit_turn_started()：发出轮次开始事件
            → capture_step_context()：捕获当前配置和执行环境
            → build_world_state_for_step()：准备运行状态视图
            → 选择压缩策略
```

`submit()` 在操作入队后返回操作标识，不等待 Session 完成路由或压缩。后台任务可以在提交方法返回前开始执行，操作提交与事件读取是两条独立路径。

`handlers::compact()` 与 `spawn_task()` 都会检查并结束旧任务，不能将两次调用理解为要执行两轮压缩。

## 3. 压缩策略选择与执行

`CompactTask.run()` 首先检查 TokenBudget；未启用时，再检查模型服务的远端压缩能力。

```text
TokenBudget 已启用
  → compact_token_budget::run_manual_compact_task()

TokenBudget 未启用，远端支持 V2
  → compact_remote_v2::run_remote_compact_task()

TokenBudget 未启用，远端不支持压缩
  → compact::run_compact_task()
```

### 3.1 客户端摘要压缩

```text
compact::run_compact_task()
  → run_compact_task_inner()
    → run_pre_compact_hooks()
    → run_compact_task_inner_impl()
      → 发出压缩项开始事件
      → 获取历史，加入压缩提示词
      → 构造摘要 Prompt
      → drain_to_completed()
        → ModelClientSession.stream()
        → 接收模型响应直至请求完成
      → 提取摘要
      → build_compacted_history()
      → 保留必要消息并重新组装运行上下文
      → Session.replace_compacted_history()
      → recompute_token_usage()
      → 发出压缩项完成事件
    → 成功时 run_post_compact_hooks()
```

客户端负责组织摘要请求和重建历史，摘要仍由模型生成。请求失败时可能重试；压缩请求自身超出上下文容量时，会尝试裁剪旧内容后重新请求。

### 3.2 远端压缩

```text
compact_remote_v2::run_remote_compact_task()
  → run_remote_compact_task_inner()
    → run_pre_compact_hooks()
    → 必要时 async_compaction::wait_for_pending()
      → 等待已有后台候选结果
    → installation::run_compaction()
      → build_and_install_compaction()
        ├─ 有适用候选：复用结果
        └─ 无候选：run_remote_compact_v2_attempt()
                    → 整理请求历史、构造压缩 Prompt
                    → run_remote_compaction_request_v2()
                    → 获取远端压缩结果
        → build_v2_compacted_history()
        → 组装保留内容与运行上下文
        → Session.replace_compacted_history()
        → recompute_token_usage()
        → 完成压缩项事件
    → 成功时 run_post_compact_hooks()
```

手动远端压缩可以复用此前自动生成的后台候选。候选失效或应用后仍超预算时，可能退回前台请求；模型兼容性等问题也可能触发回退，因此不保证只有一次模型请求。

### 3.3 TokenBudget 新窗口重建

```text
compact_token_budget::run_manual_compact_task()
  → compact_token_budget::run_compact_task_inner()
    → run_pre_compact_hooks()
    → 发出压缩项开始事件
    → Session.start_new_context_window()
      → 按配置保留部分客户端内容
      → 分配新的上下文窗口
      → assemble_initial_context()：重建初始上下文
      → Session.replace_compacted_history()
      → recompute_token_usage()
    → 发出压缩项完成事件
    → run_post_compact_hooks()
```

这个分支不请求模型生成摘要，但沿用压缩的 Hooks 和事件生命周期。这里的 `run_compact_task_inner()` 与客户端摘要模块中的同名函数不是同一个方法。

## 4. 共用的上下文替换

三种策略成功应用结果时，都会进入 `Session.replace_compacted_history()`：

```text
Session.replace_compacted_history()
  → 后台候选前缀替换时，校验历史版本和认证归属等条件
  → 生成缺失的消息标识，保留应继续存在的历史内容
  → 替换内存中模型可见的历史
  → 更新参考轮次上下文、运行状态基线等状态
  → 构造 CompactedItem 等恢复检查点记录
  → persist_rollout_items()：记录压缩结果和相关上下文
  → 标记后续会话启动 Hook 的压缩来源
```

具体策略负责生成新历史，替换方法负责将结果应用到 Session，并记录后续恢复所需的信息。前缀候选校验失败时可以拒绝应用；后续行为由外层回退策略决定。

模型用量的重新计算由各策略在应用结果后调用，不是所有步骤都封装在 `replace_compacted_history()` 内。

## 5. 任务收尾与结果回传

正常执行完成后，压缩任务返回任务调度层：

```text
CompactTask.run() 返回
  → 后台任务包装层尝试 flush_rollout()
  → Session.on_task_finished()
    → 处理任务结果与生命周期回调
    → 发出 TurnComplete
    → 清理当前活动轮次
        ↓
Session 事件通道
  → CodexThread.next_event()
  → 客户端转换并展示执行结果
```

轮次开始、压缩项开始、压缩项完成和轮次结束是不同层级的事件。压缩项完成后还可能执行压缩后 Hook；整轮是否结束，应以本轮的终态事件为准。

取消、Hook 停止或执行失败会进入对应的中断、错误处理分支。压缩失败时可能先发出错误事件，再完成任务收尾；不能仅凭提交成功或收到轮次结束事件就认定压缩成功，需要同时检查错误与终态信息。

## 6. 原仓库源码定位

以下路径相对于参考的 Codex 仓库根目录。

| 路径 | 主要职责 |
| --- | --- |
| codex-rs/core/src/codex_thread.rs | submit 与对外事件读取。 |
| codex-rs/core/src/session/mod.rs | SessionIo、上下文替换、新窗口重建与持久化协调。 |
| codex-rs/core/src/session/handlers.rs | submission_loop 与手动压缩操作分发。 |
| codex-rs/core/src/tasks/mod.rs | spawn_task、start_task、取消与任务收尾。 |
| codex-rs/core/src/tasks/compact.rs | CompactTask.run 与压缩策略选择。 |
| codex-rs/core/src/compact.rs | 客户端摘要请求与历史重建。 |
| codex-rs/core/src/compact_remote_v2.rs | 远端压缩入口、Hooks 与压缩请求。 |
| codex-rs/core/src/compact_remote_v2_attempt.rs | 构造前台或后台远端请求。 |
| codex-rs/core/src/compact_remote_v2_installation.rs | 整理和应用远端结果，处理候选与回退。 |
| codex-rs/core/src/compact_token_budget.rs | TokenBudget 手动压缩工作流程。 |
