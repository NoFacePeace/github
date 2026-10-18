# comparable 约束

`comparable` 是 Go 预声明的约束，可用于需要 `==`、`!=` 或 map 键的类型参数。它不保证支持 `<`、`>` 等大小比较。

## 1. 基本用法

```go
func Equal[T comparable](a, b T) bool {
    return a == b
}

type Set[T comparable] map[T]struct{}
```

## 2. 常见类型

| 类型 | 能否满足约束 |
| --- | --- |
| 布尔、整数、浮点数、复数、字符串 | 可以 |
| 指针、channel | 可以 |
| 数组 | 元素类型可比较时可以 |
| 结构体 | 所有字段类型可比较时可以 |
| 切片、map、函数 | 不可以 |

切片、map 和函数虽然可以与 nil 比较，但不能相互进行一般的相等比较，因此不满足 comparable。

## 3. 接口类型的特殊情况

Go 1.20 起，any 等可比较的接口类型可以满足 comparable 约束，但这不保证其所有动态值在运行时都可比较：

```go
// 可以编译，但执行时 panic：接口内装的是不可比较的切片。
// Equal[any]([]int{1}, []int{1})
```

使用接口类型作为 map 键也有类似风险。comparable 是约束，不是用于声明普通值的类型。
