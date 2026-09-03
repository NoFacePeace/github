# string

## 1. 创建

### 1.1 零值声明

```go
var s string
```

字符串的零值是空字符串 `""`。

### 1.2 字符串字面量

```go
interpreted := "hello\nworld"
raw := `hello\nworld`
```

字符串字面量包括解释型字面量和原始字面量，其字节数据通常存放在可执行文件的只读数据段。

### 1.3 从已有字符串赋值

```go
source := "hello"
target := source
```

赋值只复制字符串描述符，不复制底层字节数据。

### 1.4 截取已有字符串

```go
source := "hello"
sub := source[1:4]
```

字符串切片通常只创建新的字符串描述符，并与原字符串共享底层字节数据。

### 1.5 字符串拼接

```go
constant := "hello" + "world"
dynamic := source + target
```

常量拼接可以在编译期完成；运行时拼接通常需要分配新的存储并复制字符串内容。

### 1.6 类型转换

```go
fromBytes := string([]byte{'h', 'i'})
fromRunes := string([]rune{'h', 'i'})
fromRune := string('A')
```

- `[]byte` 转换为 `string` 时通常复制字节。
- `[]rune` 转换为 `string` 时会将 rune 编码为 UTF-8。
- 整数或 rune 转换为 `string` 时生成对应 Unicode 码点的 UTF-8 编码。

编译器在部分临时使用场景中可以消除复制。

### 1.7 通过函数或构建器生成

```go
formatted := fmt.Sprintf("%d", 10)

var builder strings.Builder
builder.WriteString("hello")
built := builder.String()
```

格式化函数、字符串构建器以及 I/O 或编码相关函数也可以返回字符串，其底层存储取决于具体实现和逃逸分析。

### 1.8 使用 unsafe 构造

```go
data := []byte("hello")
s := unsafe.String(unsafe.SliceData(data), len(data))
```

`unsafe.String` 直接引用指定地址的字节，不执行常规复制。调用者必须保证底层数据在字符串存活期间有效且不被修改。

### 1.9 底层数据分配

| 创建方式 | 通常是否创建新的底层字节数据 |
| --- | --- |
| 零值 | 否 |
| 字符串字面量 | 编译期生成静态数据 |
| 已有字符串赋值 | 否 |
| 字符串切片 | 否 |
| 运行时拼接 | 是 |
| `[]byte`、`[]rune` 转换 | 通常是 |
| 函数或构建器 | 取决于具体实现 |
| `unsafe.String` | 否 |

## 2. 底层数据结构

Go Runtime 在 `src/runtime/string.go` 中使用以下结构表示字符串：

```go
type stringStruct struct {
	str unsafe.Pointer
	len int
}
```

字符串描述符由两个字段组成：

- `str` 指向字符串的第一个字节。
- `len` 表示字符串包含的字节数。

```text
string
┌──────────────┐
│ str ─────────┼──→ 连续的字节数据
│ len          │
└──────────────┘
```

字符串描述符占用两个机器字，在 64 位平台上通常为 16 字节。底层数据不要求以 `\0` 结尾，也不保证是合法的 UTF-8。

与 slice 不同，字符串描述符没有 `cap` 字段，语言也不允许修改其底层字节。字符串赋值只复制描述符，不复制底层数据；多个字符串可以安全地共享同一段不可变字节。

`str` 指向的数据可能位于可执行文件的只读数据段、goroutine 栈或堆中，具体位置取决于字符串的创建方式和逃逸分析结果。
