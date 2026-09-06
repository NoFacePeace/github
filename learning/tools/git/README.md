# Git 学习

记录 Git 的日常操作、分支管理、对象模型和自动化机制。

## 1. Git Hook

### 1.1 什么是 Hook

Git hook 是 Git 在特定操作阶段自动调用的程序，可用于检查提交内容、校验提交信息或执行提交后的自动化操作。

Hook 由 Git 触发，与是否使用 Codex 无关。Codex hook 则由工具调用、会话等事件触发，两者属于不同的机制。

### 1.2 常用 Hook

| 固定名称 | 触发时机 | 常见用途 |
| --- | --- | --- |
| `pre-commit` | 创建提交之前 | 检查格式、运行必要检查；非零退出可阻止提交 |
| `commit-msg` | 提交信息准备好后、创建提交之前 | 校验提交信息；非零退出可阻止提交 |
| `post-commit` | 提交成功之后 | 通知、自动推送；不能撤销已经完成的提交 |
| `pre-push` | 推送过程中、更新远端引用之前 | 推送前检查；非零退出可阻止推送 |

这些文件名由 Git 规定。脚本内容可以自定义，但不能把 `post-commit` 改名为 `post-commit.sh` 并期待 Git 自动识别。

### 1.3 存放位置与执行权限

普通仓库默认从 `.git/hooks/` 查找 hook。脚本需要有执行权限，例如：

```sh
chmod +x .git/hooks/post-commit
```

`.sample` 文件是示例，不会按对应 hook 自动执行。可以通过 `core.hooksPath` 配置其他 hook 目录：

```sh
# 查看是否配置了自定义目录及配置来源
git config --show-origin --get core.hooksPath
```

`.git/hooks/` 中的文件不会被 Git 跟踪，也不会随 clone 自动安装。关联 worktree 通常共享主仓库的 hooks；不要假设每个 worktree 都有独立的 `.git/hooks/` 目录。

## 2. 实践：提交后自动推送

### 2.1 当前配置

- Hook：`post-commit`。
- 安装位置：`/Users/haotao.chen/Desktop/repositories/github/.git/hooks/post-commit`。
- 推送远端：`origin`，地址为 `git@github.com:NoFacePeace/github.git`。
- 触发流程：提交成功 → Git 调用 hook → 推送当前分支。

### 2.2 脚本内容

下面记录当前已安装的脚本，便于理解和恢复配置。修改本笔记不会自动修改实际 hook。

```sh
#!/bin/sh
# Push commits only to the explicitly approved repository.
branch=$(git symbolic-ref --quiet --short HEAD) || {
    echo '[post-commit] Detached HEAD; skipping automatic push.' >&2
    exit 0
}
expected='git@github.com:NoFacePeace/github.git'
urls=$(git remote get-url --push --all origin 2>/dev/null)
if [ "$urls" != "$expected" ]; then
    echo '[post-commit] origin push URL differs from the approved repository; skipping push.' >&2
    exit 0
fi
remote=$(git config --get "branch.$branch.remote")
target=$(git config --get "branch.$branch.merge")
if [ "$remote" = origin ] && [ -n "$target" ]; then
    git push origin "HEAD:$target"
else
    git push --set-upstream origin "HEAD:refs/heads/$branch"
fi
if [ "$?" -ne 0 ]; then
    echo '[post-commit] Automatic push failed. Your commit is saved locally; resolve the error and retry git push.' >&2
fi
exit 0
```

### 2.3 执行逻辑

1. 获取当前分支；如果处于 detached HEAD 状态，则跳过推送。
2. 检查 `origin` 的全部推送地址，仅在地址与配置的 GitHub 仓库完全一致时继续。
3. 如果当前分支配置的上游远端是 `origin`，则推送到对应的上游分支。
4. 否则推送到 `origin` 的同名分支，并建立上游跟踪关系。
5. 推送失败时显示提示，已创建的本地提交仍然保留。

这是普通推送，不包含强制推送。它会推送当前分支上远端尚未拥有的历史，因此可能一次推送多个提交。脚本同步执行，提交命令会等待推送结束。

### 2.4 验证与故障处理

```sh
# 在当前仓库根目录检查语法和执行权限
sh -n .git/hooks/post-commit
test -x .git/hooks/post-commit

# 检查本地工作区、分支和上游关系
git status --short
git branch -vv

# 查看推送目的地
git remote get-url --push --all origin
```

推送失败时，先根据错误检查网络、SSH 认证或远端分支状态；确认当前分支的上游配置后，可手动执行 `git push`。不要因为 hook 推送失败而重复创建同一份改动的提交。

### 2.5 停用方式

在当前普通仓库中，移除脚本的执行权限即可停用，恢复执行权限即可启用：

```sh
# 停用
chmod -x .git/hooks/post-commit

# 启用
chmod +x .git/hooks/post-commit
```
