# Codex Rules

`.rules` 使用 Starlark 表达命令规则，控制命令在沙箱外执行。它不是任意工具或网络资源的统一规则格式。目前为实验性功能。依据 OpenAI 官方文档整理，核对日期：2026-10-22。

## 1. 存储与加载

用户规则通常位于 `~/.codex/rules/default.rules`；项目规则位于 `.codex/rules/`，要求项目配置层受信任。修改后重启 Codex。

## 2. 命令匹配

```python
prefix_rule(
    pattern = ["gh", "pr", "view"],
    decision = "prompt",
)
```

匹配程序及参数前缀。`allow` 允许沙箱外执行，`prompt` 请求审批，`forbidden` 拒绝；多条匹配取最严格结果。复杂 Shell 不保证拆分检查，规则也不能识别脚本内部的所有行为。

## 3. 验证规则

```sh
codex execpolicy check --pretty \
  --rules ~/.codex/rules/default.rules \
  -- gh pr view 123
```

该命令检查匹配结果，不执行目标命令。

## 4. 官方文档

- [Rules：格式、匹配、加载及验证](https://learn.chatgpt.com/docs/agent-configuration/rules)
