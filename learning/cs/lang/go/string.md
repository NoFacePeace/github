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

## 3. 比较

字符串支持 `==`、`!=`、`<`、`<=`、`>` 和 `>=`。比较对象是字符串中的字节序列，而不是字符串描述符的地址，也不会按语言习惯、Unicode 规范化形式或字符数量比较。

### 3.1 相等比较

#### 3.1.1 相等条件

两个字符串相等，当且仅当长度相同且对应位置的每个字节都相同。底层数据地址是否相同不影响结果。

```go
a := "hello"
b := string([]byte{'h', 'e', 'l', 'l', 'o'})
fmt.Println(a == b) // true：地址不同，字节序列相同
fmt.Println("\u00e9" == "e\u0301") // false：显示相近，但 UTF-8 字节序列不同
```

#### 3.1.2 实现

编译器在 `src/cmd/compile/internal/walk/compare.go` 中将一般的 `s == t` 降低为：

```text
len(s) == len(t) && memequal(s.ptr, t.ptr, len(s))
```

长度不同会通过短路直接得出结果；长度相同才比较底层字节。`memequal` 由 `src/internal/bytealg/equal_*.s` 等架构相关实现提供，通常先判断数据地址是否相同，再按机器字或 SIMD 块比较，并在发现差异时提前返回。

#### 3.1.3 编译器优化

与短字符串常量比较时，编译器还可能直接生成长度检查和定宽加载比较，省去 `memequal` 调用。例如 `s == "go"` 可被降低为 `len(s) == 2` 加一次两字节比较；具体阈值取决于目标架构。

### 3.2 有序比较

#### 3.2.1 比较规则

有序比较从头逐字节比较：第一个不同字节决定结果；若公共前缀完全相同，较短的字符串更小。

```go
fmt.Println("Go" < "go") // true：'G' 的字节值小于 'g'
fmt.Println("go" < "golang") // true：公共前缀相同，短字符串更小
fmt.Println(string([]byte{0xff}) < "\u00e9") // false：string 可以包含非法 UTF-8
```

#### 3.2.2 实现

`<`、`<=`、`>` 和 `>=` 通常被降低为调用 `runtime.cmpstring`，再将其返回值与 `0` 比较：

```text
runtime.cmpstring(s, t) < 0
```

`runtime.cmpstring` 的实现位于 `src/internal/bytealg/compare_*.s` 或 `compare_generic.go`。它只扫描两者公共长度内的字节；若没有差异，再比较长度。以 amd64 为例，较长输入会使用向量指令成块查找首个不同字节。

#### 3.2.3 标准库接口

`strings.Compare` 返回 `-1`、`0` 或 `1`，适合需要三路比较结果的接口；普通条件判断直接使用比较运算符更清晰。

### 3.3 性能与并发

字符串比较不分配内存，时间复杂度最坏为 `O(min(len(s), len(t)))`；共同前缀越长，需要读取的内存越多。字符串不可变，因此多个 goroutine 可以安全地同时比较同一字符串，但若通过 `unsafe.String` 引用仍在被修改的字节，比较会产生数据竞争并破坏字符串不可变假设。
