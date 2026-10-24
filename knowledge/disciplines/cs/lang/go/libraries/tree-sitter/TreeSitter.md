# Tree-sitter Go 使用

`github.com/tree-sitter/go-tree-sitter` 是 Tree-sitter 的 Go 绑定。Go 程序可以在进程内解析源码，获取语法树并提取代码结构。跨语言的输入输出、增量解析与能力边界见 [Tree-sitter 原理](../../../../compiler/parsing/TreeSitter.md)。

## 1. 依赖与构建

需要分别引入绑定和目标语言 grammar：绑定提供 Parser、Tree 和 Node 等接口，grammar 定义所解析语言的语法。下面使用 Go grammar 解析 Go 源码；绑定也能配合其他 grammar 解析其他语言。

```bash
go get github.com/tree-sitter/go-tree-sitter
go get github.com/tree-sitter/tree-sitter-go
```

这条接入路径依赖 CGO 和 C 编译工具链。构建环境需启用 CGO，并保留依赖模块中的 C 源文件；交叉编译时还需匹配目标平台的 C 工具链。

实际项目应固定兼容版本，并提交 `go.mod`、`go.sum`。以下示例采用 `go-tree-sitter` v0.24.0 的 API 形式。

## 2. 解析源码

```go
package main

import (
    "fmt"

    tree_sitter "github.com/tree-sitter/go-tree-sitter"
    tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func main() {
    source := []byte(`package demo

func CreateOrder(id string) {
    SaveOrder(id)
}
`)

    parser := tree_sitter.NewParser()
    defer parser.Close()

    language := tree_sitter.NewLanguage(tree_sitter_go.Language())
    if err := parser.SetLanguage(language); err != nil {
        panic(err)
    }

    tree := parser.Parse(source, nil)
    if tree == nil {
        panic("parser returned no tree")
    }
    defer tree.Close()

    root := tree.RootNode()
    if root.HasError() {
        panic("source contains syntax errors")
    }
    fmt.Println(root.ToSexp())
    visit(root, source)
}

func visit(node *tree_sitter.Node, source []byte) {
    if node.Kind() == "function_declaration" {
        name := node.ChildByFieldName("name")
        if name != nil {
            fmt.Printf("function=%s line=%d\n",
                name.Utf8Text(source), node.StartPosition().Row+1)
            fmt.Println(string(source[node.StartByte():node.EndByte()]))
        }
    }
    for i := uint(0); i < node.NamedChildCount(); i++ {
        visit(node.NamedChild(i), source)
    }
}
```

`Parse(source, nil)` 执行不提供旧树的解析。源码中的调用目标不需要实际存在，才能构建语法树；名称绑定和类型检查是后续分析工作。

## 3. 节点遍历与源码提取

| API | 用途 |
| --- | --- |
| `Kind()` | 获取节点类型，例如 `function_declaration`、`call_expression` |
| `ChildByFieldName("name")` | 按 grammar 定义的字段名取得子节点 |
| `NamedChildCount()` / `NamedChild()` | 遍历具名节点，跳过部分标点等匿名节点 |
| `StartByte()` / `EndByte()` | 取得字节范围，按左闭右开区间截取原始源码 |
| `StartPosition()` / `EndPosition()` | 获取从 0 开始的行号和字节列号 |
| `Utf8Text(source)` | 利用原始源码读取节点对应文本 |

节点类型和字段名由 grammar 决定。Go 方法使用 `method_declaration`，上面只匹配 `function_declaration` 的示例不会提取方法。中文等多字节字符下，字节偏移和字节列号不能直接当作字符序号。

## 4. 资源生命周期与错误处理

- Parser 和 Tree 分配底层 C 内存，应显式调用 `Close()`，不能只依赖 Go GC。
- 如果使用 TreeCursor、Query、QueryCursor 等其他持有 C 资源的对象，也需按 API 要求释放。
- Node 依附于 Tree；Tree 关闭后不要继续访问其节点。需要长期保存时，复制名称、位置和源码等普通 Go 数据。
- Tree-sitter 支持错误恢复，返回非空 Tree 不代表语法正确。严格分析可以检查 `HasError()`，编辑器则可能接受部分语法树。
- 增量解析需先将编辑范围应用到旧树，再把旧树和新源码交给解析器；仅重复调用 `Parse(source, nil)` 不会复用旧树。

## 5. 参考资料

- [Go 绑定源码与使用说明](https://github.com/tree-sitter/go-tree-sitter)。
- [Go 绑定 API](https://pkg.go.dev/github.com/tree-sitter/go-tree-sitter)。
- [Go grammar](https://github.com/tree-sitter/tree-sitter-go)。
