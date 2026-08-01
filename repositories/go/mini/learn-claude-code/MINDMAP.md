# learn-claude-code 思维导图

> 按 20 章逐节整理的思维导图。每章列出关键概念、数据结构/机制与一句话定位。
> 源仓库：<https://github.com/shareAI-lab/learn-claude-code>

---

## 一图总览（文本树形）

```
learn-claude-code
│
├─ s01 Agent Loop ─「一个循环 + Bash 就够了」
│   ├─ 核心机制：messages / while True / stop_reason
│   ├─ 实现：单条 while 循环 + 只有一个 Bash 工具
│   └─ 定位：一切 agent 的地基——先让循环转起来
│
├─ s02 Tool Use ─「加工具即加一个 handler」
│   ├─ 核心机制：dispatchTool / switch 分发 / 并发
│   ├─ 设计：循环不动，工具注册进 dispatch map
│   └─ 定位：把"加工具"变成"加一段配置"，可扩展
│
├─ s03 Permission ─「先立边界，再给自由」
│   ├─ 核心机制：denyList / permissionRule / checkPermission
│   ├─ 三态判断：可执行 / 必须停下（denyList）/ 需审批
│   └─ 定位：在能力扩开前，先决定"能否跑"
│
├─ s04 Hooks ─「围绕循环打补丁，不重写循环」
│   ├─ 核心机制：hookEvent / hooks map / triggerHooks
│   ├─ 扩展点：UserPromptSubmit / PreToolUse / PostToolUse / Stop
│   └─ 定位：不动主干、横向织入副作用的扩展模式
│
├─ s05 TodoWrite ─「没有计划的 agent 会漂移」
│   ├─ 核心机制：todoItem / roundsSinceTodo / nagRounds
│   ├─ 流程：先写待办列表，再逐步执行更新
│   └─ 定位：给循环装上"目标感"，避免漫无目的
│
├─ s06 Subagent ─「大任务拆小，子任务用干净上下文」
│   ├─ 核心机制：spawnSubagent / fresh messages[] / subMaxTurns
│   ├─ 模式：子 agent 干杂活，只回传结果摘要
│   └─ 定位：保护主上下文，把污染留在子进程
│
├─ s07 Skill Loading ─「知识按需加载，而非一次性灌输」
│   ├─ 核心机制：skillRegistry / listSkills / loadSkill
│   ├─ 流程：先列技能清单进 SYSTEM，用到才展开
│   └─ 定位：技能目录化 + 懒加载，避免上下文爆炸
│
├─ s08 Context Compact ─「上下文总会满，要留出空间」★复杂章
│   ├─ 多层策略：snipCompact / microCompact / toolResultBudget
│   ├─ 预算控制：contextLimit / autoCompact / reactiveCompact
│   └─ 定位：长对话续命术，把"满"推迟、把"废料"剪掉
│
├─ s09 Memory ─「记住重要的，忘掉无关的」
│   ├─ 三子系统：selection / extraction / consolidation
│   ├─ 选择：挑哪些值得记；抽取：提炼事实
│   └─ 定位：跨会话的长期记忆，区别于上下文压缩
│
├─ s10 System Prompt ─「提示词是装配出来的，不是写死的」
│   ├─ 核心机制：运行时组装 / 分段拼接
│   ├─ 方式：按需加载各 section 再拼接
│   └─ 定位：系统提示也变成可维护的运行时产物
│
├─ s11 Error Recovery ─「错误不是终点，是重试的起点」
│   ├─ 策略：升级 token / 回退模型 / 重试
│   ├─ 后路：重试 / 腾空间 / 走替代路径
│   └─ 定位：失败可恢复，agent 不会卡死在一处
│
├─ s12 Task System ─「大目标拆小任务，有序、落盘」
│   ├─ 核心机制：TaskRecord / blockedBy / 磁盘持久化
│   ├─ 结构：任务图 + 依赖关系 + 文件落地
│   └─ 定位：为多 agent 协作铺路的任务基座
│
├─ s13 Background Tasks ─「慢活进后台，agent 继续想」
│   ├─ 核心机制：线程执行 / 通知队列
│   ├─ 模式：后台跑命令，完成时注入通知
│   └─ 定位：耗时操作不阻塞主循环
│
├─ s14 Cron Scheduler ─「到点自动跑，无需人工催」
│   ├─ 核心机制：持久化调度 / 会话级触发
│   ├─ 模式：按时间自动触发任务
│   └─ 定位：让 agent 拥有"定时自主行动"能力
│
├─ s15 Agent Teams ─「太大就交给队友」
│   ├─ 核心机制：MessageBus / 收件箱 / 权限上浮
│   ├─ 模式：常驻队友 + 异步邮箱
│   └─ 定位：从单兵到团队的协作骨架
│
├─ s16 Team Protocols ─「队友需要共同的通信规则」
│   ├─ 核心机制：关停握手 / 计划审批
│   ├─ 规范：固定 请求-应答 格式
│   └─ 定位：给团队协作定"语法/握手"
│
├─ s17 Autonomous Agents ─「队友自己看板领活干」
│   ├─ 核心机制：空闲循环 / 自动领取 / 自组织
│   ├─ 模式：无领导指派，任务看板自取
│   └─ 定位：团队自驱，不再靠人逐个派单
│
├─ s18 Worktree Isolation ─「各干各目录，互不干扰」
│   ├─ 核心机制：WorktreeRecord / 任务-目录绑定
│   ├─ 映射：任务持有目标，工作树持有目录，用 ID 绑定
│   └─ 定位：多 agent 并行改代码时的物理隔离
│
├─ s19 MCP Plugin ─「能力不够，用 MCP 外接」
│   ├─ 核心机制：多传输 / 通道路由 / 工具池装配
│   ├─ 方式：外部工具接入同一工具池
│   └─ 定位：通过协议外挂无限扩展能力
│
└─ s20 Comprehensive Agent ─「机制有很多，循环只有一个」
    ├─ 核心：所有机制围绕同一条循环整合
    ├─ 整合：工具/权限/Hook/记忆/任务/团队/MCP 全收口
    └─ 定位：终点章——把前 19 章拼成一个完整 harness
```
