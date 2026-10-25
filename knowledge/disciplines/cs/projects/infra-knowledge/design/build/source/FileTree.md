# 文件树

文件树描述仓库在指定 Commit 下的文件路径、类型和对象标识，供源码准备进行[筛选](filter/Filter.md)，再读取选中文件的内容。获取文件树时不需要先检出整个工作目录。

## 1. 输入与输出

- 输入：仓库地址、固定 Commit 和访问凭据。
- 输出：该版本的文件条目集合，不包含文件正文。

```text
模式     类型    对象 ID     路径
100644   blob    abc123…     src/main.go
100644   blob    def456…     config.yaml
120000   blob    789abc…     current
160000   commit  456def…     modules/example
```

对象 ID 用于后续定位文件内容。普通文件、可执行文件、符号链接和子模块需要根据模式分别处理。

## 2. 通过 Git 获取

在构建机器上的临时裸仓库中执行命令。远程服务支持部分获取时，可以先获取指定版本的树与提交对象，将文件内容按需获取。

```bash
git init --bare /tmp/source-tree
git -C /tmp/source-tree remote add origin <repository-url>
git -C /tmp/source-tree config remote.origin.promisor true
git -C /tmp/source-tree config remote.origin.partialclonefilter blob:none
git -C /tmp/source-tree fetch --no-tags --depth=1 --filter=blob:none origin <commit-sha>
git -C /tmp/source-tree ls-tree -r -z --full-tree <commit-sha>
```

`-r` 递归列出文件，`-z` 使用 NUL 字符分隔记录，避免文件名中的空格等字符影响解析。命令在构建机器本地执行，获取对象时会访问远程仓库。

```go
输出 := 执行Git("ls-tree", "-r", "-z", "--full-tree", commit)
for _, 记录 := range 按NUL字符拆分(输出) {
    条目 := 解析模式类型对象ID和路径(记录)
    校验路径(条目.Path)
    文件清单 = append(文件清单, 条目)
}
```

路径筛选通过后，可使用 `git cat-file blob <object-id>` 获取选中文件的内容；本地缺失的对象由 Git 按需从远程获取。

## 3. 通过仓库 API 获取

远程服务不支持所需的 Git 部分获取能力时，可以通过仓库 API 分页查询固定版本的递归文件树。

```go
for 有下一页 {
    当前页 := 查询文件树(仓库, commit, 递归=true, 页码)
    for _, 条目 := range 当前页.条目 {
        校验路径并检查重复(条目.Path)
        if 条目.类型 == "目录" {
            continue
        }
        文件清单 = append(文件清单, 条目)
    }
    页码 = 当前页.下一页
}
```

后续只对通过筛选的文件调用内容读取接口。文件树与内容读取始终使用同一个 Commit，避免分支移动导致输入版本混杂。
