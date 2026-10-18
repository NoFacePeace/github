# any 约束

`any` 是 `interface{}` 的别名。用作类型约束时允许任意类型，但不保证这些类型支持比较、算术或特定方法。

## 1. 基本用法

```go
func Identity[T any](value T) T {
    return value
}

var number = Identity(123)   // int
var text = Identity("abc")   // string
```

输入和输出都使用 T，因此它们保留同一种具体类型。`[T any]` 不会将参数的静态类型改成接口 any。

## 2. 允许与不保证的操作

可以声明 T 的变量、赋值、传递和返回 T，也可以构建 `[]T` 等类型。下面的泛型函数返回任意 T 的零值：

```go
func Zero[T any]() T {
    var value T
    return value
}
```

不能直接对两个 T 使用 `==`、`+` 或调用未声明的方法，因为任意类型并不都支持这些操作。也不能直接对类型参数值做类型断言；如需运行时类型检查，可以先转换为接口值，例如 `any(value)`。

## 3. 适用场景

适合保存或传递元素的容器，以及不依赖元素特定运算的算法。需要比较、排序或方法调用时，应增加相应约束。
