# 联合约束

约束中的 `|` 表示类型集的并集。它允许一个类型参数取多类类型，但不表示一个值可以同时包含多种类型。

## 1. 基本用法

```go
type Number interface {
    ~int | ~int64 | ~float64
}

func Add[T Number](a, b T) T {
    return a + b
}
```

每次实例化仍只有一个 T。例如 T 为 int64 时，两个参数和返回值都是 int64，不会因为约束列了 float64 就自动混合不同的已类型化参数。

## 2. 操作必须对所有成员有效

`int | string` 都支持 `+`，但含义分别是加法和字符串拼接。`int | bool` 则不能直接使用 `+`，因为 bool 不支持它。

允许的类型范围越广，能统一执行的操作不一定越多。

## 3. 并集与交集

```go
type NamedNumber interface {
    ~int | ~int64
    String() string
}
```

`~int | ~int64` 内部取并集，再与方法要求取交集。满足的类型既要有相应底层类型，也要实现方法。

## 4. 常见限制

多个非接口类型项的类型集不能重叠，例如 `int | ~int` 非法，因为 `~int` 已包含 int。

多项联合不能包含 comparable，也不能包含声明方法或嵌入 comparable 的接口。因此，不应写 `comparable | ~int` 或将 `fmt.Stringer` 作为多项联合中的分支。

带联合项的非基本接口只能用于约束，不能用作普通变量类型；它不是可以直接承载 int 或 string 的联合值类型。
