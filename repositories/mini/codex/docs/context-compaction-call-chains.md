# 上下文压缩入口与调用链

本文记录 Codex 原仓库的上下文压缩入口、策略选择、共用实现和结果应用流程。当前 Go 项目已定义接口与数据类型骨架，尚未实现压缩；文中的 Rust 方法用于后续设计参考。

Go 接口按职责定义在 [local.go](../core/compaction/local.go)、[remote.go](../core/compaction/remote.go)、[token_budget.go](../core/compaction/token_budget.go) 和 [history.go](../core/compaction/history.go) 中，请求与结果类型见 [compaction.go](../core/compaction/compaction.go)。这些是本项目的接口设计，参数类型经过简化，不是原 Rust 签名的逐项转换。

## 1. 入口与策略

按触发方式分为手动和自动两类。自动压缩包含前台执行与后台预计算两条路径，后台预计算属于自动压缩的优化机制。

| 路径 | 入口 | 是否启动独立 CompactTask | 对当前工作的影响 |
| --- | --- | --- | --- |
| 手动压缩 | 调用方提交 Op::Compact，例如 /compact。 | 是。 | 替换当前任务，压缩结束后不自动恢复被替换的任务。 |
| 自动前台压缩 | 普通任务内部检查上下文用量、模型切换等条件。 | 否。 | 执行中压缩成功后继续当前任务；也可在轮次结束时整理上下文。 |
| 自动后台预计算 | async_compaction::maybe_start() 检查开关与阈值。 | 否，但会启动后台异步工作。 | 先生成候选结果，后续在合适的检查点校验并应用。 |

触发方式与压缩策略是两个维度。手动和自动前台入口采用相同的策略选择顺序：

```text
是否启用 TokenBudget？
  ├─ 是：重建新的上下文窗口，不请求模型生成摘要
  └─ 否：模型服务是否支持远端压缩 V2？
      ├─ 是：远端压缩
      └─ 否：客户端组织模型摘要请求
```

“客户端摘要”指客户端组织摘要生成和历史重建，摘要仍由模型生成。

## 2. 手动压缩调用链

从操作提交到结果应用、事件回传的完整过程见[手动上下文压缩完整调用链](manual-context-compaction-call-chain.md)。

```text
调用方执行 /compact 或直接提交压缩操作
  → CodexThread.submit(Op::Compact)
  → SessionIo：操作入队
  → submission_loop()
  → handlers::compact()
    → abort_all_tasks(TurnAbortReason::Replaced)
    → new_turn_with_default_settings()
    → spawn_task(..., CompactTask)
      → CompactTask.run()
        → 发出轮次开始事件
        → 捕获配置、环境与当前上下文视图
        → 选择压缩策略
          ├─ compact::run_compact_task()
          ├─ compact_remote_v2::run_remote_compact_task()
          └─ compact_token_budget::run_manual_compact_task()
        → 返回，由任务调度层完成本轮收尾
```

这是独立的压缩工作流程，不会在压缩后自动接着执行被替换的普通任务。客户端是否允许执行中发起手动压缩，由相应界面或协议入口决定；以上描述的是 Core 收到操作后的行为。

## 3. 自动前台压缩调用链

```text
RegularTask.run()
  → RegularTask.run_turns()
  → run_turn()
    ├─ 请求前
    │   → run_pre_sampling_compact()
    │     ├─ maybe_run_previous_model_inline_compact()
    │     │   → 检查模型切换后的上下文容量或压缩兼容性变化
    │     └─ 检查当前用量与后台候选结果
    ├─ 执行中
    │   → 检查后续工作、上下文阈值或新窗口请求
    └─ 轮次结束时
        → 检查结束后的压缩阈值或后台候选结果
              ↓
        run_auto_compact()
          → 选择压缩策略
            ├─ compact::run_inline_auto_compact_task()
            ├─ compact_remote_v2::run_inline_remote_auto_compact_task()
            └─ compact_token_budget::run_inline_auto_compact_task()
```

请求前、执行中和轮次结束时是自动压缩的不同阶段，另有特定上下文超限恢复分支调用 `run_auto_compact()`。这些调用按各自条件触发，并非每轮必定全部执行。

执行中的自动压缩通常使用当前 Turn 的上下文；成功后继续 agent loop。轮次结束阶段的压缩用于后续请求，不表示重新开始已经完成的工作。失败或取消时，由调用阶段决定停止、报错或保留已完成结果。

`session/turn.rs` 将远端自动入口导入为 `run_inline_remote_auto_compact_task_v2`，它是同一函数的别名。

## 4. 手动与自动共用的策略实现

### 4.1 客户端摘要压缩

```text
手动：compact::run_compact_task()
自动：compact::run_inline_auto_compact_task()
                   ↓
      compact::run_compact_task_inner()
        → 执行压缩前 Hook
        → run_compact_task_inner_impl()
          → 获取历史并构造摘要请求
          → 请求模型生成摘要
          → 组合摘要、保留消息和必要运行上下文
          → Session.replace_compacted_history()
          → 重新计算上下文用量并报告压缩项完成
        → 成功时执行压缩后 Hook
```

两种包装方法传入不同的触发来源、原因和阶段信息，核心摘要与历史重建逻辑共用。

### 4.2 远端压缩

```text
手动：compact_remote_v2::run_remote_compact_task()
自动：compact_remote_v2::run_inline_remote_auto_compact_task()
                   ↓
      run_remote_compact_task_inner()
        → 执行压缩前 Hook
        → 必要时等待已有后台候选结果
        → installation::run_compaction()
          → build_and_install_compaction()
            ├─ 有候选结果：复用候选并校验适用性
            └─ 无候选结果：run_remote_compact_v2_attempt()
                           → 请求远端压缩结果
            → 整理保留内容并重建上下文
            → Session.replace_compacted_history()
            → 重新计算上下文用量
        → 完成事件及成功后的压缩后 Hook 等收尾
```

候选结果不可应用时，是否发起前台请求由回退策略决定；应用候选后仍超过预算时，也可能再进行一次前台压缩。模型兼容性变化等情况还可能触发模型回退，不保证每次只发出一个请求。

### 4.3 TokenBudget 新窗口重建

```text
手动：compact_token_budget::run_manual_compact_task()
自动：compact_token_budget::run_inline_auto_compact_task()
                   ↓
      compact_token_budget::run_compact_task_inner()
        → 执行压缩前 Hook
        → 发出压缩项开始事件
        → Session.start_new_context_window()
          → 按配置保留部分客户端内容
          → 重建初始上下文与运行状态
          → Session.replace_compacted_history()
          → 重新计算上下文用量
        → 发出压缩项完成事件
        → 执行压缩后 Hook
```

该策略跳过模型摘要请求，通过建立新上下文窗口管理预算。不同模块中存在同名的 `run_compact_task_inner()`，它们是不同函数。

## 5. 自动后台预计算调用链

### 5.1 生成候选结果

```text
run_turn()
  → async_compaction::maybe_start()
    → 检查功能开关、服务能力、阈值和会话状态
    → capture_background_input()：捕获历史快照与相关状态
    → 启动后台异步工作
      → run_background_compaction()
        → run_remote_compact_v2_attempt()
        → 返回 BackgroundCompactionCandidate
      → 通过通道保存候选结果
```

后台预计算要求远端压缩 V2 能力及相应功能开关，TokenBudget 模式下不启用。它提前生成压缩结果，不直接替换正在变化的历史；失败的后台尝试会被丢弃。

### 5.2 取出并应用候选结果

```text
自动检查点
  → async_compaction::take_ready()
  → run_auto_compact(candidate)
  → 远端自动压缩包装方法
                   ↓
      run_remote_compact_task_inner()
        → installation::run_compaction()
          → build_and_install_compaction()
            → 校验模型及压缩兼容性
            → Session.replace_compacted_history()
              → 校验历史版本、前缀和认证归属
              → 替换捕获的前缀，保留之后新增的历史
```

远端压缩流程在没有显式候选且需要前台回退时，也可以通过 `wait_for_pending()` 等待已有后台结果，因此手动远端压缩同样可能复用后台预计算。

## 6. 共用边界

| 层级 | 是否共用 |
| --- | --- |
| 最外层触发入口 | 手动和自动不同；后台预计算是自动机制的另一条路径。 |
| 策略选择 | 手动任务与自动入口分别选择，但遵循相同的分支顺序。 |
| 同一种策略的核心方法 | 手动与自动共用内部实现。 |
| 成功应用结果 | 三种策略最终都进入 Session.replace_compacted_history()。 |
| 完成后的控制流 | 手动任务收尾，执行中的自动压缩继续原任务，后台预计算等待后续采用。 |

`replace_compacted_history()` 是更新 Session 中模型可见历史及相关检查点的共用方法，不负责统一完成所有压缩策略的结果生成。后台候选校验失败时可以不发生替换。

## 7. 原仓库源码定位

以下路径相对于参考的 Codex 仓库根目录，具体函数可能随源码版本调整。

| 路径 | 职责 |
| --- | --- |
| codex-rs/core/src/session/handlers.rs | 接收手动压缩操作，替换任务并启动 CompactTask。 |
| codex-rs/core/src/tasks/compact.rs | 手动压缩工作流程与策略选择。 |
| codex-rs/core/src/session/turn.rs | 自动触发检查、run_auto_compact 及策略选择。 |
| codex-rs/core/src/compact.rs | 客户端摘要压缩的手动、自动包装与共用实现。 |
| codex-rs/core/src/compact_remote_v2.rs | 远端压缩包装、Hooks 与候选获取。 |
| codex-rs/core/src/compact_remote_v2_attempt.rs | 前台和后台远端压缩请求及候选生成。 |
| codex-rs/core/src/compact_remote_v2_installation.rs | 整理与应用远端结果、处理候选及回退。 |
| codex-rs/core/src/compact_token_budget.rs | TokenBudget 模式下的新窗口重建工作流程。 |
| codex-rs/core/src/session/async_compaction.rs | 后台预计算调度、候选读取与等待。 |
| codex-rs/core/src/session/mod.rs | start_new_context_window 与 replace_compacted_history。 |
