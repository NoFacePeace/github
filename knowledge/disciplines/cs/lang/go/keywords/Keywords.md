# Keywords

Go 关键字是语言保留的词，用于声明、控制流和并发等语法，不能作为普通标识符使用。按关键字组织专题，可以在同一篇文档中介绍语法、使用场景、编译器处理及运行时实现。

## 1. 关键字分类

Go 有 25 个关键字，以下按主要用途分类，每个关键字列出一次。

| 分类 | 关键字 | 主要用途 |
| --- | --- | --- |
| 包与导入 | `package`、`import` | 声明包与导入依赖 |
| 声明 | `const`、`var`、`type`、`func` | 声明常量、变量、类型和函数 |
| 类型构造 | `struct`、`interface`、`map`、`chan` | 构造相应类型 |
| 条件与分支 | `if`、`else`、`switch`、`case`、`default` | 条件判断与分支选择，后两者也用于 select |
| 循环 | `for`、`range` | 循环与遍历 |
| 执行跳转 | `break`、`continue`、`goto`、`fallthrough`、`return` | 终止、跳转或继续执行 |
| 并发与延迟调用 | `go`、`select`、`defer` | 启动 goroutine、选择通信操作和安排延迟调用 |

分类用于导航，不代表每个关键字只承担一种作用。例如 `func` 也用于函数类型和函数字面量，`range` 的可遍历对象随 Go 版本演进。

## 2. 关键字与预声明标识符

`int`、`string`、`bool`、`any`、`nil`、`true`、`false`、`make`、`new`、`len` 等不是关键字，而是预声明标识符。它们与关键字不同，可以在内层作用域被同名声明遮蔽，但通常应避免这样做，以免影响可读性。

`context.Context` 是标准库接口类型，也不是关键字。

## 3. 用法与实现

关键字首先由编译器识别和处理，具体执行不一定依赖 runtime。

- `if`、`for` 等通常转换为分支、跳转及相关机器指令。
- `go` 的执行涉及创建与调度 goroutine。
- `select` 由编译器根据分支形式转换，一般多分支路径调用运行时的 `selectgo()`。
- `defer` 的实现受编译器优化和使用场景影响，可能通过开放编码或运行时辅助路径完成。

每个专题可以按语法、示例、行为规则、编译器处理和源码入口展开，结合实际版本说明实现细节。

## 4. 专题文档

- [Select](Select.md)：多路 channel 通信选择、编译器处理与运行时实现。

## 5. 参考资料

- [Keywords](https://go.dev/ref/spec#Keywords)：语言规范中的关键字列表。
- [Predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers)：预声明标识符。

[返回 Go 目录](../Go.md)
