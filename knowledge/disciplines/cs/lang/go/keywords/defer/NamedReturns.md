# defer 与具名返回值

具名返回值是函数中可访问的结果变量。return 先设置结果变量，随后执行的 defer 可以继续修改它，调用方最终取得修改后的结果。

## 1. 修改具名结果

```go
func Example() (result int) {
    defer func() {
        result = 2
    }()
    return 1
}
// 返回 2。
```

虽然写的是 `return 1`，执行时仍先将 1 赋给 result。defer 修改的正是这个返回结果变量，因此最终返回 2。

## 2. 裸 return

```go
func Bare() (result int) {
    result = 1
    defer func() {
        result++
    }()
    return
}
// 返回 2。
```

裸 return 不显式提供表达式，使用当前结果变量。它不会跳过 defer，也不会提前冻结结果变量的值。

## 3. 多个 defer 的修改顺序

```go
func Order() (result int) {
    defer func() { result *= 10 }()
    defer func() { result += 2 }()
    return 1
}
// 先设置为 1，再加 2，最后乘 10，返回 30。
```

结果取决于实际登记顺序。只有访问并修改同一个结果变量才会改变返回结果；登记时保存一份参数副本，不会自动修改外部结果变量。

## 4. 参数副本与外部结果变量

```go
func Copy() (result int) {
    defer func(value int) {
        value = 99
    }(result)
    return 1
}
// 返回 1。
```

登记 defer 时 result 的初始值为 0，这个值复制给延迟调用的参数。修改参数 value 不会修改 result。若确实要影响返回值，可在闭包中直接修改 result，或按明确的指针协议操作它。

## 5. 使用注意

具名返回值常用于在 defer 中补充错误处理，但修改应保持清晰，避免无意覆盖原错误。若 defer 引发未恢复的 panic，则不会正常向调用方返回结果。
