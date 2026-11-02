# Tree-sitter

Tree-sitter 是解析器生成工具和增量解析库。它依据语言语法将源码转换为具体语法树（CST），并能在源码编辑后复用旧树进行增量解析。它属于传统程序分析基础设施，解析过程不依赖大模型。

## 1. 输入与输出

| 项目 | 内容 |
| --- | --- |
| 输入 | 源码文本，以及对应语言的 grammar；增量解析还需旧树和编辑信息 |
| 处理 | 识别声明、表达式、参数和控制结构等语法结构 |
| 输出 | 语法树，节点包含类型、父子关系、字节范围和行列位置，并可检查语法错误或缺失节点 |

以 Go 代码为例：

```go
func CreateOrder(id string) {
    SaveOrder(id)
}
```

省略部分节点后的树结构示意如下：

```text
source_file
└── function_declaration
    ├── name: CreateOrder
    ├── parameters
    │   └── parameter_declaration
    │       ├── name: id
    │       └── type: string
    └── body
        └── call_expression
            ├── function: SaveOrder
            └── arguments: id
```

下游分析器可以遍历这些节点，提取函数、调用表达式和源码位置，再按项目规则组织成代码组件与关系。

## 2. 典型用途与边界

典型用途包括语法高亮、结构化代码选择、源码查询、静态分析及为 AI 工具准备代码上下文。

Tree-sitter 提供语法结构，不自动完成完整的类型检查、跨文件符号绑定或运行时调用图，也不直接生成业务文档。例如，识别 `SaveOrder(id)` 的语法后，仍需下游分析确定它关联哪个声明。

它支持错误恢复，存在语法错误时可能仍返回语法树；消费方可以自行选择跳过错误文件或使用部分结果。

## 3. Go 接入

运行库核心使用 C11 实现，提供 CLI 和多种语言绑定。Go 程序通常将绑定与目标语言 grammar 作为依赖，通过 CGO 在进程内解析，无需独立解析服务。

```go
import (
    tree_sitter "github.com/tree-sitter/go-tree-sitter"
    tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)
```

`go-tree-sitter` 提供解析器和树的访问接口，`tree-sitter-go` 提供 Go 语言的语法规则。使用时还需按绑定 API 管理 Parser 和 Tree 的生命周期。

## 4. 参考资料

- [Tree-sitter 官方介绍](https://tree-sitter.github.io/tree-sitter/)。
- [Tree-sitter Go 绑定](https://github.com/tree-sitter/go-tree-sitter)。
- [Go 语言 grammar](https://github.com/tree-sitter/tree-sitter-go)。
