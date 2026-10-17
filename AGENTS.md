# AGENTS

## 1. Commit Attribution

When creating git commits, always append this trailer:
Co-authored-by: Codex <noreply@openai.com>

## 2. 提交信息约定

- 提交标题使用英文，简洁概括本次改动。
- 提交正文使用中文，简要说明本次提交包含的具体改动。

## 3. 目录命名约定

- 新建目录时，目录名使用英文。

## 4. 文档标题约定

- 文档一级标题作为文档名称，不添加编号。
- 从二级标题开始，按层级使用有序编号，例如 `## 1. 基础概念`、`### 1.1 核心概念`、`#### 1.1.1 示例`。
- 同级标题按出现顺序连续编号；新增、删除或调整章节顺序时，同步更新相关编号。

## 5. 目录简介

- [knowledge](knowledge/Knowledge.md)：个人知识库，按学科知识（disciplines）、工具使用（tools）、生活知识（life）、想法与创意（ideas）分类整理 Markdown 笔记，各目录通过对应文档记录知识内容并组织子文档链接。

### 5.1 知识库文件链接约定

- 在 `knowledge` 下创建文件时，默认同步在相应的目录对应文档中添加相对链接。
- 目录对应文档是与目录主题对应的文档，可以保存该主题的知识内容，也可以组织子文档链接，不要求它是专门的索引文件。
- 创建当前目录的对应文档时，默认由上级目录的对应文档链接它。例如，创建 `acm/graph/Graph.md` 时，在 `acm/ACM.md` 中添加链接。知识库根文档 `knowledge/Knowledge.md` 无需按此规则向上添加链接。
- 创建当前目录下的其他文档时，默认由当前目录的对应文档链接它。例如，创建 `acm/graph/MinimumSwapsToSortByDigitSum.md` 时，在 `acm/graph/Graph.md` 中添加链接。
- 识别目录对应文档时，优先沿用现有命名和组织方式，不仅依赖文件名与目录名是否完全一致。例如，`graph` 对应 `Graph.md`，`data-structures` 对应 `DataStructures.md`。
- 找不到目录对应文档时，先检查现有目录结构及相关文档的链接，确认组织方式后再处理，不猜测文件名或随意新建目录对应文档。
- 添加链接前检查是否已有指向同一文件的链接，避免重复；使用相对路径，并在完成后检查链接目标存在。
