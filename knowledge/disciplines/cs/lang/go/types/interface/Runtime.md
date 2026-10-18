# 接口运行时检查

接口值可以承载不同的动态类型。类型断言和类型分支根据实际动态类型进行判断，从而决定能否作为目标接口使用，或能否取出目标具体类型的值。

## 1. 判断动态类型是否实现接口

```go
type Reader interface {
    Read(p []byte) (int, error)
}

func TryReader(value any) (Reader, bool) {
    r, ok := value.(Reader)
    return r, ok
}
```

调用 `TryReader(123)` 时，接口内的动态类型是 int，它没有 Read 方法，所以返回 nil 和 false。如果传入的动态类型实现了 Reader，则断言成功，返回的接口承载原来的动态类型和值。

判断的是实际类型是否满足方法要求，不会尝试调用一次 Read 来验证，也不会运行时扫描 Go 源码。

## 2. 断言到具体类型

```go
var value any = int64(123)

n, ok := value.(int64) // 成功：n 为 123，ok 为 true
other, matched := value.(int) // 失败：other 为 0，matched 为 false
```

断言目标为具体类型时，要求动态类型与目标类型相同。int64 和 int 即使数值相同，也不会自动转换；类型断言不是数值类型转换。

## 3. 什么时候 panic

| 写法 | 断言成功 | 断言失败 |
| --- | --- | --- |
| `r, ok := value.(Reader)` | 原接口值与 true | nil 与 false |
| `r := value.(Reader)` | 返回目标接口值 | panic |
| `n, ok := value.(int)` | 目标值与 true | int 零值 0 与 false |
| `n := value.(int)` | 返回目标值 | panic |

使用双返回值形式可以处理断言失败，但不负责捕获后续方法调用中产生的 panic。

## 4. nil 接口与 nil 指针

nil 接口没有动态类型，对它进行类型断言一定失败：

```go
var value any

// r, ok := value.(Reader) // 返回 nil、false
// r := value.(Reader)     // panic
```

但接口内装有 nil 指针时，仍然有动态类型。如果 `*File` 实现 Reader，下面的断言可以成功：

```go
// 假设 *File 已实现 Reader
// var file *File
// var value any = file
// r, ok := value.(Reader) // ok 为 true，r 不等于 nil
```

这里并没有创建非 nil 的 File。后续调用方法是否安全，取决于方法是否正确处理 nil 接收者。

## 5. 类型分支

```go
func Kind(value any) string {
    switch value.(type) {
    case nil:
        return "nil"
    case Reader:
        return "reader"
    case string:
        return "string"
    default:
        return "other"
    }
}
```

接口 case 检查动态类型是否实现接口，具体类型 case 检查动态类型是否相同。多个 case 都能匹配时，选择源码中首先匹配的分支；没有匹配时进入 default，或直接结束 switch，不会仅因没有匹配而 panic。

## 6. 类型元数据与缓存

运行时利用编译器生成的类型和方法元数据完成动态检查。目标是非空接口时，相关实现可能查找或构造 itab，记录目标接口与具体类型的对应关系及方法入口。

运行时具有相关缓存，编译器也可能直接确定或优化部分检查。因此，并不是每次断言都重新逐个匹配全部方法；具体路径以版本和编译结果为准。

## 7. 源码入口

- `src/runtime/iface.go`：itab 查找、接口断言及断言失败处理等辅助逻辑。
- `src/internal/abi`：接口方法与类型元数据的内部表示。
- `src/cmd/compile/internal`：类型断言、类型分支的检查与代码生成。
