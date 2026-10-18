# Go 泛型

Go 泛型通过类型参数，让同一份函数或类型定义适用于多种类型，并保留编译期类型检查。Go 从 1.18 开始支持泛型，常用于容器、集合和通用算法。

## 1. 类型参数

类型参数写在函数名或类型名后的方括号中。下面的函数要求参数和返回值具有同一种类型：

```go
func Identity[T any](x T) T {
    return x
}

func example() {
    n := Identity(123)       // 推断 T 为 int，n 为 int
    s := Identity("hello")   // 推断 T 为 string，s 为 string
    _ = n
    _ = s
}
```

| 写法 | 含义 |
| --- | --- |
| `T` | 类型参数名，可以由定义者命名 |
| `any` | 类型约束，允许任意类型 |
| `x T` | 参数 x 的类型是 T |
| 返回值 `T` | 返回值必须符合同一个 T |

`T` 表示调用时确定的类型，不是可以在函数内部任意切换类型的变量。

## 2. 与 any 的区别

`any` 是 `interface{}` 的别名。普通函数使用 `any` 时，可以接收不同类型的值，但不能仅凭签名要求输入与输出类型相同：

```go
func IdentityAny(x any) any {
    return x
}

func exampleAny() {
    result := IdentityAny(123) // 静态类型为 any
    n := result.(int)         // 要作为 int 使用，需要类型断言
    _ = n
}
```

这个函数即使改为返回字符串，也满足原来的签名。泛型版本的输入和输出都使用 T，则会在编译期约束它们的类型关系：传入 int，返回值就是 int；传入 string，返回值就是 string。

`[T any]` 中的 `any` 表示允许哪些类型作为 T，不会把参数 x 的静态类型变成接口 `any`。

## 3. 泛型函数

下面的函数返回切片首元素，并通过布尔值区分空切片：

```go
func First[T any](items []T) (T, bool) {
    if len(items) == 0 {
        var zero T
        return zero, false
    }
    return items[0], true
}
```

`var zero T` 获取 T 的零值。例如，T 为 int 时是 0，为 string 时是空字符串，为指针类型时是 nil。并非所有 T 都能直接用 nil 作为返回值。

## 4. 类型约束与类型集

各类约束、类型集组合及使用限制参见[泛型约束](constraints/Constraints.md)。

### 4.1 约束决定允许的操作

泛型函数中的操作必须对约束允许的所有类型都有效。`any` 不保证支持加法或相等比较；如果需要 `==`，可以使用 `comparable`：

```go
func Equal[T comparable](a, b T) bool {
    return a == b
}
```

int、string 等类型满足该约束，切片、map 和函数类型不满足。接口类型作为类型实参时还存在细节：Go 1.20 起，any 可以满足 comparable 约束，但接口值在运行时包含不可比较的动态值时，比较仍可能 panic。

### 4.2 使用接口声明约束

约束通过接口描述允许的类型集，也可以要求方法。下面只允许底层类型为 int 或 int64 的类型：

```go
type Integer interface {
    ~int | ~int64
}

func Add[T Integer](a, b T) T {
    return a + b
}

type Count int

var total = Add(Count(1), Count(2)) // 结果为 Count
```

`|` 表示类型集的并集；`~int` 表示底层类型为 int 的所有类型，因此包含自定义的 Count。单独写 int 则不包含 `type Count int` 这样的新定义类型。

带有这类类型项的约束接口只能用于类型约束，不能像普通接口一样直接声明变量来存放值。

## 5. 类型推断与显式指定

编译器可以根据调用参数等信息推断类型参数：

```go
var inferred = Identity(123)
var explicit = Identity[int](123)
```

两者的结果类型都是 int。如果信息不足，则需要显式指定类型，例如：

```go
func Zero[T any]() T {
    var zero T
    return zero
}

var number = Zero[int]()
```

不能简单写 `number := Zero()`，再期待编译器根据接收变量推断 T。

## 6. 泛型类型与方法

泛型类型可以让容器保留元素的具体类型：

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    last := len(s.items) - 1
    v := s.items[last]
    var zero T
    s.items[last] = zero
    s.items = s.items[:last]
    return v, true
}

var numbers Stack[int]
var names Stack[string]
```

方法通过接收者声明使用类型参数。Go 不支持方法再额外声明一组独立的类型参数；需要这种能力时，可以使用泛型函数。泛型类型使用时需要实例化，例如 Stack[int]，不能省略为 Stack。

## 7. 使用场景

| 需求 | 常见选择 |
| --- | --- |
| 相同算法用于多个类型，并需要保持输入输出类型关系 | 泛型函数 |
| 栈、集合等容器需要保存具体元素类型 | 泛型类型 |
| 不同实现提供同一组行为，运行时统一调用 | 普通接口 |
| 业务逻辑只处理一种明确类型 | 普通函数或类型 |

泛型与接口可以配合使用。泛型保证的类型关系不等于业务逻辑一定正确，也不意味着一定比普通接口更快；性能应依据具体实现和测量结果判断。
