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

编译器在部分临时使用场景中可以消除复制，具体实现见第 4 节。

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

## 4. 类型转换

字符串与字节切片之间的转换保留原始字节；字符串与 rune 切片或整数之间的转换需要执行 UTF-8 解码或编码。涉及切片的常规转换会生成独立的结果，以维持字符串不可变而切片可变的语义。

### 4.1 字符串类型之间的转换

底层类型为 `string` 的类型之间相互转换时，字节内容不变，只需复制字符串描述符，不会复制底层数据：

```go
type Name string

n := Name("go")
s := string(n)
```

这类转换在编译器内部属于无操作转换，不需要调用 runtime，也不会分配内存。

### 4.2 `[]byte` 转换为 `string`

#### 4.2.1 转换语义

`string(b)` 生成内容为 `b` 当前字节序列的字符串。常规转换必须复制数据，因此之后修改 `b` 不会改变字符串：

```go
b := []byte("go")
s := string(b)
b[0] = 'n'
fmt.Println(s) // go
```

#### 4.2.2 实现

编译器在 `src/cmd/compile/internal/walk/convert.go` 中将转换降低为 `runtime.slicebytetostring`。该函数对空切片直接返回空字符串，对单字节结果复用静态表，其他情况选择栈缓冲区或堆内存，并调用 `memmove` 复制字节。

若结果不逃逸且长度不超过 64 字节，编译器会传入栈上临时缓冲区。此时仍会复制，只是避免了堆分配；更长或逃逸的结果在堆上分配。

#### 4.2.3 编译器优化

当转换结果仅用于字符串比较、字符串拼接或 `map` 查找等不会保留并修改原切片的场景时，编译器可以使用临时转换 `OBYTES2STRTMP`，让字符串直接引用切片数据，避免分配和复制。该结果不能超出编译器证明安全的使用范围。

### 4.3 `string` 转换为 `[]byte`

`[]byte(s)` 创建可修改的字节切片，因此常规转换需要复制字符串内容。编译器通常将其降低为 `runtime.stringtoslicebyte`；结果不逃逸且长度不超过 64 字节时可以使用栈上缓冲区，否则分配新的底层数组。

字符串常量转换可以直接初始化定长字节数组。对于 `for range []byte(s)` 等编译器能够证明切片不会被修改的临时场景，内部的 `OSTR2BYTESTMP` 可以直接读取字符串数据，省去复制；该优化不会改变普通 `[]byte(s)` 返回独立可变切片的语义。

### 4.4 `[]rune` 转换为 `string`

`string(rs)` 将每个 rune 编码为 UTF-8，非法 Unicode 码点会编码为 `U+FFFD`。编译器将转换降低为 `runtime.slicerunetostring`：第一次遍历计算编码后的字节数并分配空间，第二次遍历写入编码结果。

结果不逃逸且较小时可以使用 64 字节的栈上缓冲区。转换时间与 rune 数量及编码结果长度成正比。

### 4.5 `string` 转换为 `[]rune`

`[]rune(s)` 按 UTF-8 解码字符串；每段非法 UTF-8 编码按 `range` 的规则产生 `U+FFFD`。`runtime.stringtoslicerune` 先遍历一次统计 rune 数量，再分配切片并进行第二次遍历写入 rune。

结果不逃逸且不超过 32 个 rune 时可以使用栈上缓冲区。所得切片拥有独立存储，修改它不会影响原字符串。

### 4.6 整数转换为 `string`

整数转换生成该 Unicode 码点的 UTF-8 编码，而不是整数的十进制文本：

```go
fmt.Println(string(65)) // A
```

编译器将其降低为 `runtime.intstring`。有效码点最多编码为 4 个字节；负数、超出 Unicode 范围或代理区码点转换为 `U+FFFD`。结果不逃逸时使用 4 字节栈上缓冲区，需要十进制格式应使用 `strconv.Itoa` 或 `strconv.FormatInt`。

### 4.7 复制与分配总结

| 转换 | 常规是否复制 | 不逃逸时的典型存储 |
| --- | --- | --- |
| 字符串类型 → 字符串类型 | 否 | 复用底层数据 |
| `[]byte` → `string` | 是 | 不超过 64 字节时使用栈缓冲区 |
| `string` → `[]byte` | 是 | 不超过 64 字节时使用栈缓冲区 |
| `[]rune` → `string` | 编码到新存储 | 较小时使用 64 字节栈缓冲区 |
| `string` → `[]rune` | 解码到新存储 | 不超过 32 个 rune 时使用栈缓冲区 |
| 整数 → `string` | 编码到新存储 | 4 字节栈缓冲区 |

“不发生堆分配”不等于“不复制”：栈缓冲区优化仍会复制或编码数据；只有编译器证明生命周期和可变性安全的临时转换才可能真正复用底层数据。
