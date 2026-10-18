# 接口

接口通过方法要求和类型集描述一组类型。普通接口变量可以承载满足接口要求的值，使代码依赖行为而非具体实现；泛型约束还可以用接口限定类型参数的范围。

## 1. 声明与隐式实现

```go
type Speaker interface {
    Speak() string
}

type User struct {
    Name string
}

func (u User) Speak() string {
    return u.Name
}

func Say(s Speaker) string {
    return s.Speak()
}

var message = Say(User{Name: "Alice"})
```

User 的方法集满足 Speaker，因此实现该接口，不需要显式写 implements。方法名、参数和返回值必须符合要求，仅有名称相同的方法并不够。

赋值与方法集的静态判断参见[接口编译期检查](Compile.md)。

## 2. 静态类型与动态类型

```go
var s Speaker = User{Name: "Alice"}
```

| 概念 | 示例 |
| --- | --- |
| 变量的静态类型 | Speaker |
| 接口承载值的动态类型 | User |
| 动态值 | `User{Name: "Alice"}` |

通过 s 可以直接调用接口声明的 Speak 方法，不能因为动态值是 User 就直接访问 `s.Name`。需要使用具体类型的字段或额外方法时，可以进行类型断言。

## 3. 方法集与指针接收者

对于普通定义类型 T，T 的方法集包含接收者为 T 的方法，`*T` 的方法集还包含接收者为 `*T` 的方法。

如果 Speak 仅定义在 `*User` 上，满足 Speaker 的是 `*User`，而不是 User：

```go
// 假设 Speak 使用指针接收者：func (u *User) Speak() string
// var s Speaker = &User{Name: "Alice"} // 可以
// var s Speaker = User{Name: "Alice"}  // 不满足接口
```

普通方法调用对可寻址变量自动取地址的语法便利，不会改变该类型是否实现接口的判断。

## 4. 空接口与 any

`any` 是 `interface{}` 的别名，没有方法要求，可以承载任意类型的值。

```go
var value any = 123
value = "hello"
```

变量的静态类型始终是 any，只是动态类型和值发生变化。使用具体类型能力前，需要进行相应的类型检查或转换。

## 5. 类型断言与类型分支

```go
func Describe(value any) string {
    if text, ok := value.(string); ok {
        return text
    }

    switch v := value.(type) {
    case int:
        _ = v // 此分支中 v 为 int
        return "integer"
    case nil:
        return "nil"
    default:
        return "other"
    }
}
```

`v, ok := value.(T)` 断言失败时返回 T 的零值和 false；不带 ok 的 `value.(T)` 断言失败时会 panic。断言目标也可以是另一个接口，此时检查动态类型是否实现目标接口。

动态类型判断、断言失败与 panic 的具体行为参见[接口运行时检查](Runtime.md)。

## 6. nil 接口与带类型的 nil

```go
var empty any
var pointer *User
var wrapped any = pointer

// empty == nil   为 true
// wrapped == nil 为 false
```

接口值只有在动态类型和值都不存在时才等于 nil。wrapped 虽然承载 nil 指针，仍然具有动态类型 `*User`，所以不等于 nil。

这也是返回 error 时常见的问题：把具体错误类型的 nil 指针装入 error 接口，不等于直接返回 nil。对带 nil 指针的接口调用方法时是否 panic，取决于方法实现是否安全处理 nil 接收者。

## 7. 接口比较

接口可以进行相等比较，但需要考虑动态类型和值：

- 动态类型不同时，两个非 nil 接口值不相等。
- 动态类型相同且可比较时，根据动态值比较结果判断。
- 动态类型相同但不可比较时，例如两个接口都装有 `[]int`，比较会 panic。

接口与 nil 比较不会因为动态值是切片等不可比较类型而 panic。使用接口作为 map 键时，也必须保证实际动态值可比较。

## 8. 接口嵌入与泛型约束

接口可以嵌入其他接口，组合方法要求：

```go
type Reader interface {
    Read(p []byte) (int, error)
}

type ReadCloser interface {
    Reader
    Close() error
}
```

只包含方法要求的基本接口既可以声明普通变量，也可以作为泛型约束。

泛型引入的 `~int`、`int | string` 等类型项，以及 comparable，可以进一步限定类型集。包含这些元素的非基本接口只能用于约束，不能直接作为承载值的普通变量类型。

## 9. 运行时表示

从理解接口行为的角度，可以将接口值看作动态类型信息与数据的组合。常见实现对空接口和非空接口使用不同表示：

| 表示 | 主要信息 |
| --- | --- |
| 空接口 | 动态类型信息与数据 |
| 非空接口 | 包含类型和方法分派信息的 itab，以及数据 |

具体数据布局、是否需要分配和方法调用能否被优化，由编译器、运行时及版本决定。接口转换不一定分配堆内存，不能简单认为“使用接口就会逃逸”。

## 10. 源码入口

- `src/runtime/iface.go`：接口转换、类型断言等辅助逻辑。
- `src/runtime/runtime2.go`：接口相关内部表示，具体组织随版本变化。
- `src/internal/abi/iface.go`：itab 等 ABI 结构。
- `src/cmd/compile/internal`：接口类型检查、转换与调用的代码生成。
