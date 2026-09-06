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

## 3. 使用 git-filter-repo 精简历史

### 3.1 工具定位

`git-filter-repo` 是用于重写 Git 历史的独立开源命令行工具，需要单独安装。它不是 Git 内置命令，也不是 GitHub 开发的工具。

Git 官方文档推荐使用它替代旧的 `git filter-branch`，GitHub 官方文档也使用它进行历史清理和仓库拆分。

- [项目仓库：newren/git-filter-repo](https://github.com/newren/git-filter-repo)
- [Git 官方文档：git-filter-branch](https://git-scm.com/docs/git-filter-branch)
- [GitHub 官方文档：拆分子目录为独立仓库](https://docs.github.com/en/get-started/using-git/splitting-a-subfolder-out-into-a-new-repository)

### 3.2 安装与调用

macOS 可以通过 Homebrew 安装：

```sh
brew install git-filter-repo
git filter-repo --version
```

执行 `git filter-repo` 时，Git 会查找名为 `git-filter-repo` 的可执行程序。也可以直接使用 Python 执行下载的工具脚本；本次清理采用临时下载的方式，没有全局安装。

### 3.3 为什么删除文件后仓库仍然很大

普通删除并提交，只会让文件从新的快照中消失，旧提交仍引用它的内容。完整克隆需要获取相关历史，因此工作区大小与 Git 历史大小可能相差很大。

`git-filter-repo` 会过滤历史提交中的指定路径，重建提交及父子关系。受影响的提交及其后续提交 ID 会变化。重新打包并回收不再被引用的对象后，仓库才能释放对应空间。

### 3.4 按路径清理

以下为本次在独立副本中执行的清理示例：

```sh
# 第一轮：从历史中移除资料目录
git filter-repo --path files/ --invert-paths

# 第二轮：移除 AI 资料和编译产物，保留 DataHub 源码
git filter-repo \
  --path ai/ \
  --path repositories/go/datahub/datahub \
  --invert-paths
```

- `--path`：指定仓库根目录下的相对路径，可重复传入。
- `--invert-paths`：删除指定路径，保留其他内容。
- 路径过滤不会自动跟随历史重命名；清理前需要确认旧路径是否也应包含。
- `repositories/go/datahub/datahub` 是历史编译产物，当前源码位于 `repositories/go/projects/datahub/`，不在删除范围内。

### 3.5 备份、验证与同步流程

1. 备份实际资料，并使用 `git bundle create <备份路径> --all` 保存所有引用可达的历史。Bundle 不包含未提交文件、忽略文件和 hooks，这些需要单独备份。
2. 通过 `git bundle verify <备份路径>` 验证历史备份，在独立镜像副本中执行过滤，避免影响当前工作区。
3. 对比清理前后的文件树，确认只有指定路径被移除；检查目标路径的历史是否消失，并执行 `git fsck --full` 检查对象完整性。
4. 核对远端当前提交，再使用逐分支指定旧提交 ID 的 `--force-with-lease` 替换远端历史；多个分支可配合 `--atomic` 一并更新。
5. 同步本地分支和关联 worktree。其他电脑上的旧克隆可以重新克隆，避免把旧历史重新合并回来。

历史替换需要明确选择分支和标签，不应把备份副本中的全部引用直接镜像推送到远端。独立清理副本也不安装自动推送 hook，推送在验证后单独执行。

### 3.6 本次精简结果

| 阶段 | 历史备份或精简副本体积（约） |
| --- | ---: |
| 清理前的完整历史 Bundle | 208 MB |
| 移除 `files/` 后的镜像副本 | 101 MB |
| 再移除 `ai/` 和 DataHub 编译产物后的镜像副本 | 6.9 MB |

这些数值用于观察量级变化，Bundle 与镜像目录的统计口径略有不同，不代表当前工作目录的总占用。第二轮核对了 204 个保留提交，其他文件内容及 DataHub 源码历史均保持不变。

备份位置：`/Users/haotao.chen/Desktop/repositories/github-backup-20260906/`。其中 `files/` 保存资料，`history.bundle` 保存第一轮之前的原始历史，`second-pass/` 保存第二轮备份和清理副本。

原工作仓库的 reflog 仍可能引用旧提交，所以本地 `.git` 不会立即缩小到精简副本的体积；远端旧对象的空间释放也取决于平台回收。历史过滤、更新分支引用和回收旧对象是不同步骤。
